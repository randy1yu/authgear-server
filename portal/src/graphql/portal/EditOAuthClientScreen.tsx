import cn from "classnames";
import React, { useCallback, useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import deepEqual from "deep-equal";
import produce, { createDraft } from "immer";
import { Icon, Text, Link, useTheme, Image, ImageFit } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import ScreenContent from "../../ScreenContent";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import EditOAuthClientForm, {
  getReducedClientConfig,
} from "./EditOAuthClientForm";
import {
  ApplicationType,
  OAuthClientConfig,
  PortalAPIAppConfig,
} from "../../types";
import { clearEmptyObject } from "../../util/misc";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";
import styles from "./EditOAuthClientScreen.module.css";
import Widget from "../../Widget";
import flutterIconURL from "../../images/framework_flutter.svg";
import xamarinIconURL from "../../images/framework_xamarin.svg";

interface FormState {
  publicOrigin: string;
  clients: OAuthClientConfig[];
  editedClient: OAuthClientConfig | null;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    publicOrigin: config.http?.public_origin ?? "",
    clients: config.oauth?.clients ?? [],
    editedClient: null,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.oauth ??= {};
    config.oauth.clients = currentState.clients.slice();

    const client = currentState.editedClient;
    if (client) {
      const index = config.oauth.clients.findIndex(
        (c) => c.client_id === client.client_id
      );
      if (
        index !== -1 &&
        !deepEqual(
          getReducedClientConfig(client),
          getReducedClientConfig(config.oauth.clients[index]),
          { strict: true }
        )
      ) {
        config.oauth.clients[index] = createDraft(client);
      }
    }
    clearEmptyObject(config);
  });
}

interface QuickStartFrameworkItem {
  icon: React.ReactNode;
  name: string;
  docLink: string;
}

interface QuickStartWidgetProps {
  applicationType?: ApplicationType;
}

const QuickStartWidget: React.FC<QuickStartWidgetProps> =
  function QuickStartWidget(props) {
    const { applicationType } = props;
    const { renderToString } = useContext(Context);
    const theme = useTheme();

    const items: QuickStartFrameworkItem[] = useMemo(() => {
      switch (applicationType) {
        case "spa":
          return [
            {
              icon: <i className={cn("fab", "fa-react")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.react"
              ),
              docLink: "https://docs.authgear.com/tutorials/spa/react",
            },
            {
              icon: <i className={cn("fab", "fa-vuejs")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.vue"
              ),
              docLink: "https://docs.authgear.com/get-started/website",
            },
            {
              icon: <i className={cn("fab", "fa-angular")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.angular"
              ),
              docLink: "https://docs.authgear.com/get-started/website",
            },
            {
              icon: <i className={cn("fab", "fa-js")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.other-js"
              ),
              docLink: "https://docs.authgear.com/get-started/website",
            },
          ];
        case "traditional_webapp":
          return [
            {
              icon: <Icon iconName="Globe" />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.traditional-webapp"
              ),
              docLink: "https://docs.authgear.com/get-started/website",
            },
          ];
        case "native":
          return [
            {
              icon: <i className={cn("fab", "fa-react")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.react-native"
              ),
              docLink: "https://docs.authgear.com/get-started/react-native",
            },
            {
              icon: <i className={cn("fab", "fa-apple")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.ios"
              ),
              docLink: "https://docs.authgear.com/get-started/ios",
            },
            {
              icon: <i className={cn("fab", "fa-android")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.android"
              ),
              docLink: "https://docs.authgear.com/get-started/android",
            },
            {
              icon: (
                <Image
                  src={flutterIconURL}
                  imageFit={ImageFit.contain}
                  className={styles.frameworkImage}
                />
              ),
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.flutter"
              ),
              docLink: "https://docs.authgear.com/get-started/flutter",
            },
            {
              icon: (
                <Image
                  src={xamarinIconURL}
                  imageFit={ImageFit.contain}
                  className={styles.frameworkImage}
                />
              ),
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.xamarin"
              ),
              docLink: "https://docs.authgear.com/get-started/xamarin",
            },
          ];
        default:
          return [];
      }
    }, [applicationType, renderToString]);

    if (applicationType == null) {
      return null;
    }

    return (
      <Widget>
        <div className={styles.quickStartWidget}>
          <div>
            <Icon
              className={styles.quickStartTitleIcon}
              styles={{ root: { color: theme.palette.themePrimary } }}
              iconName="Lightbulb"
            />
            <Text className={styles.quickStartTitle}>
              <FormattedMessage id="EditOAuthClientScreen.quick-start.title" />
            </Text>
          </div>
          <Text>
            <FormattedMessage id="EditOAuthClientScreen.quick-start.question" />
          </Text>
          {items.map((item, index) => (
            <Link
              key={`quick-start-${index}`}
              className={styles.quickStartItem}
              href={item.docLink}
              target="_blank"
            >
              <span className={styles.quickStartItemIcon}>{item.icon}</span>
              <Text variant="small" className={styles.quickStartItemText}>
                {item.name}
              </Text>
              <Icon
                className={styles.quickStartItemArrowIcon}
                iconName="ChevronRightSmall"
              />
            </Link>
          ))}
        </div>
      </Widget>
    );
  };

interface EditOAuthClientContentProps {
  form: AppConfigFormModel<FormState>;
  clientID: string;
}

const EditOAuthClientContent: React.FC<EditOAuthClientContentProps> =
  function EditOAuthClientContent(props) {
    const {
      clientID,
      form: { state, setState },
    } = props;

    const client =
      state.editedClient ?? state.clients.find((c) => c.client_id === clientID);

    const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
      return [
        {
          to: "./../..",
          label: (
            <FormattedMessage id="ApplicationsConfigurationScreen.title" />
          ),
        },
        {
          to: ".",
          label: client?.name ?? "",
        },
      ];
    }, [client?.name]);

    const onClientConfigChange = useCallback(
      (editedClient: OAuthClientConfig) => {
        setState((state) => ({ ...state, editedClient }));
      },
      [setState]
    );

    if (client == null) {
      return (
        <Text>
          <FormattedMessage
            id="EditOAuthClientScreen.client-not-found"
            values={{ clientID }}
          />
        </Text>
      );
    }

    return (
      <ScreenContent>
        <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
        <div className={cn(styles.widget, styles.widgetColumn)}>
          <EditOAuthClientForm
            publicOrigin={state.publicOrigin}
            clientConfig={client}
            onClientConfigChange={onClientConfigChange}
          />
        </div>
        <div className={styles.quickStartColumn}>
          <QuickStartWidget applicationType={client.x_application_type} />
        </div>
      </ScreenContent>
    );
  };

const EditOAuthClientScreen: React.FC = function EditOAuthClientScreen() {
  const { appID, clientID } = useParams() as {
    appID: string;
    clientID: string;
  };
  const form = useAppConfigForm(appID, constructFormState, constructConfig);

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <EditOAuthClientContent form={form} clientID={clientID} />
    </FormContainer>
  );
};

export default EditOAuthClientScreen;
