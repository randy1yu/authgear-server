package analytic

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type SignupCountResult struct {
	TotalCount           int
	CountByLoginID       map[string]int
	CountByOAuthProvider map[string]int
	AnonymousCount       int
}

type CountCollector struct {
	GlobalHandle  *globaldb.Handle
	GlobalDBStore *GlobalDBStore
	AppDBHandle   *appdb.Handle
	AppDBStore    *AppDBStore
	AuditDBHandle *auditdb.WriteHandle
	AuditDBStore  *AuditDBStore
	CounterStore  *ReadStoreRedis
}

func (c *CountCollector) CollectDaily(date *time.Time) (updatedCount int, err error) {
	utc := date.UTC()
	rangeFrom := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
	rangeTo := rangeFrom.AddDate(0, 0, 1)

	var appIDs []string
	err = c.GlobalHandle.WithTx(func() error {
		appIDs, err = c.GlobalDBStore.GetAppIDs()
		if err != nil {
			return fmt.Errorf("failed to fetch app ids: %w", err)
		}
		return nil
	})
	if err != nil {
		return
	}

	var counts []*Count
	for _, appID := range appIDs {
		appCounts, e := c.CollectDailyCountForApp(appID, rangeFrom, rangeTo)
		if e != nil {
			err = e
			return
		}
		counts = append(counts, appCounts...)
	}

	if len(counts) > 0 {
		err = c.AuditDBHandle.WithTx(func() error {
			// Store the counts to audit db
			err = c.AuditDBStore.CreateCounts(counts)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return 0, fmt.Errorf("failed to store count: %w", err)
		}
	}

	updatedCount = len(counts)
	return updatedCount, nil
}

func (c *CountCollector) CollectDailyCountForApp(appID string, date time.Time, nextDay time.Time) (counts []*Count, err error) {
	// Cumulative number of user count
	err = c.AppDBHandle.WithTx(func() error {
		userCount, err := c.AppDBStore.GetUserCountBeforeTime(appID, &nextDay)
		if err != nil {
			return err
		}
		if userCount == 0 {
			// no user in the app, skip the cumulative number of user count
			return nil
		}
		counts = append(counts, NewCount(
			appID,
			userCount,
			date,
			CumulativeUserCountType,
		))
		return nil
	})
	if err != nil {
		err = fmt.Errorf("failed to calculate cumulative number of user: %w", err)
		return
	}

	// Signup count
	err = c.AuditDBHandle.ReadOnly(func() error {
		signupCountResult, err := c.querySignupCount(appID, &date, &nextDay)
		if err != nil {
			err = fmt.Errorf("failed to calculate signup count: %w", err)
			return err
		}
		if signupCountResult.TotalCount == 0 {
			// no new signup for the app, skip the signup count
			return nil
		}

		counts = append(counts, NewCount(
			appID,
			signupCountResult.TotalCount,
			date,
			DailySignupCountType,
		))

		for loginIDType, count := range signupCountResult.CountByLoginID {
			counts = append(counts, NewDailySignupWithLoginID(
				appID,
				count,
				date,
				loginIDType,
			))
		}

		for provider, count := range signupCountResult.CountByOAuthProvider {
			counts = append(counts, NewDailySignupWithOAuth(
				appID,
				count,
				date,
				provider,
			))
		}

		if signupCountResult.AnonymousCount != 0 {
			counts = append(counts, NewCount(
				appID,
				signupCountResult.AnonymousCount,
				date,
				DailySignupAnonymouslyCountType,
			))
		}
		return nil
	})
	if err != nil {
		return
	}

	// Collect counts from redis
	dailyCount, err := c.CounterStore.GetDailyCountResult(config.AppID(appID), &date)
	if err != nil {
		return
	}

	if dailyCount.ActiveUser != 0 {
		counts = append(counts, NewCount(
			appID,
			dailyCount.ActiveUser,
			date,
			DailyActiveUserCountType,
		))
	}

	if dailyCount.LoginPageView != 0 {
		counts = append(counts, NewCount(
			appID,
			dailyCount.LoginPageView,
			date,
			DailyLoginPageViewCountType,
		))
	}

	if dailyCount.LoginUniquePageView != 0 {
		counts = append(counts, NewCount(
			appID,
			dailyCount.LoginUniquePageView,
			date,
			DailyLoginUniquePageViewCountType,
		))
	}

	if dailyCount.SignupPageView != 0 {
		counts = append(counts, NewCount(
			appID,
			dailyCount.SignupPageView,
			date,
			DailySignupPageViewCountType,
		))
	}

	if dailyCount.SignupUniquePageView != 0 {
		counts = append(counts, NewCount(
			appID,
			dailyCount.SignupUniquePageView,
			date,
			DailySignupUniquePageViewCountType,
		))
	}

	return
}

func (c *CountCollector) querySignupCount(appID string, rangeFrom *time.Time, rangeTo *time.Time) (*SignupCountResult, error) {
	var first uint64 = 100
	var after model.PageCursor = ""

	result := &SignupCountResult{
		CountByLoginID:       map[string]int{},
		CountByOAuthProvider: map[string]int{},
	}
	for {
		events, lastCursor, err := c.queryUserCreatedEvents(appID, rangeFrom, rangeTo, first, after)
		if err != nil {
			return nil, err
		}

		// Termination condition
		if len(events) == 0 {
			return result, nil
		}

		after = lastCursor
		for _, e := range events {
			result.TotalCount++
			payload := e.Payload.(*nonblocking.UserCreatedEventPayload)
			if len(payload.Identities) < 1 {
				log.Fatal("missing user identities")
			}
			iden := payload.Identities[0]
			switch authn.IdentityType(iden.Type) {
			case authn.IdentityTypeLoginID:
				loginIDType := iden.Claims[identity.IdentityClaimLoginIDType].(string)
				if loginIDType == "" {
					log.Fatal("missing type in login id identity claims")
				}
				result.CountByLoginID[loginIDType]++
			case authn.IdentityTypeOAuth:
				provider := iden.Claims[identity.IdentityClaimOAuthProviderType].(string)
				if provider == "" {
					log.Fatal("missing provider in oauth identity claims")
				}
				result.CountByOAuthProvider[provider]++
			case authn.IdentityTypeAnonymous:
				result.AnonymousCount++
			}
		}
	}
}

func (c *CountCollector) queryUserCreatedEvents(appID string, rangeFrom *time.Time, rangeTo *time.Time, first uint64, after model.PageCursor) (events []*event.Event, lastCursor model.PageCursor, err error) {
	options := audit.QueryPageOptions{
		RangeFrom:     rangeFrom,
		RangeTo:       rangeTo,
		ActivityTypes: []string{string(nonblocking.UserCreated)},
	}

	logs, offset, err := c.AuditDBStore.QueryPage(appID, options, graphqlutil.PageArgs{
		First: &first,
		After: graphqlutil.Cursor(after),
	})
	if err != nil {
		return
	}
	events = make([]*event.Event, len(logs))
	for i, log := range logs {
		b, e := json.Marshal(log.Data)
		if e != nil {
			err = e
			return
		}
		eventObj := event.Event{
			Payload: &nonblocking.UserCreatedEventPayload{},
		}
		e = json.Unmarshal(b, &eventObj)
		if e != nil {
			err = e
			return
		}
		events[i] = &eventObj
	}

	pageKey := db.PageKey{Offset: offset + uint64(len(logs)) - 1}
	cursor, err := pageKey.ToPageCursor()
	if err != nil {
		return
	}
	after = cursor

	return events, after, nil
}
