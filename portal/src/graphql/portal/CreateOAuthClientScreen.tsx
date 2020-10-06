import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  Callout,
  Dialog,
  DialogFooter,
  DirectionalHint,
  IconButton,
  Label,
  PrimaryButton,
  Text,
} from "@fluentui/react";
import { useNavigate, useParams } from "react-router-dom";
import produce from "immer";
import deepEqual from "deep-equal";
import { FormattedMessage } from "@oursky/react-messageformat";

import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ModifyOAuthClientForm from "./ModifyOAuthClientForm";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import { useAppConfigQuery } from "./query/appConfigQuery";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import { OAuthClientConfig, PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import { genRandomHexadecimalString } from "../../util/random";
import { copyToClipboard } from "../../util/clipboard";

import styles from "./CreateOAuthClientScreen.module.scss";

interface CreateOAuthClientFormProps {
  rawAppConfig: PortalAPIAppConfig;
}

interface CreateClientSuccessDialogProps {
  visible: boolean;
  clientId: string;
}
const CreateClientSuccessDialog: React.FC<CreateClientSuccessDialogProps> = function CreateClientSuccessDialog(
  props: CreateClientSuccessDialogProps
) {
  const { visible, clientId } = props;
  const navigate = useNavigate();

  const calloutIntervalRef = useRef<number>();
  const [isCalloutVisible, setIsCalloutVisible] = useState(false);

  useEffect(() => {
    if (calloutIntervalRef.current != null) {
      window.clearTimeout(calloutIntervalRef.current);
    }
  }, []);

  const onConfirmCreateClientSuccess = useCallback(() => {
    navigate("../");
  }, [navigate]);

  const onCopyClick = useCallback(() => {
    copyToClipboard(clientId);
    setIsCalloutVisible(true);
    calloutIntervalRef.current = window.setTimeout(() => {
      setIsCalloutVisible(false);
    }, 3000);
  }, [clientId]);

  return (
    <Dialog
      hidden={!visible}
      title={
        <FormattedMessage id="CreateOAuthClientScreen.success-dialog.title" />
      }
    >
      <Label>
        <FormattedMessage id="CreateOAuthClientScreen.success-dialog.client-id-label" />
      </Label>
      <div className={styles.dialogClientId}>
        <Text>{clientId}</Text>
        <IconButton
          onClick={onCopyClick}
          className={styles.dialogCopyIcon}
          iconProps={{ iconName: "Copy" }}
        />
      </div>
      {isCalloutVisible && (
        <Callout
          className={styles.copyButtonCallout}
          target={`.${styles.dialogCopyIcon}`}
          directionalHint={DirectionalHint.bottomLeftEdge}
        >
          <Text>
            <FormattedMessage id="CreateOAuthClientScreen.success-dialog.copied" />
          </Text>
        </Callout>
      )}
      <DialogFooter>
        <PrimaryButton onClick={onConfirmCreateClientSuccess}>
          <FormattedMessage id="done" />
        </PrimaryButton>
      </DialogFooter>
    </Dialog>
  );
};

const CreateOAuthClientForm: React.FC<CreateOAuthClientFormProps> = function CreateOAuthClientForm(
  props: CreateOAuthClientFormProps
) {
  const { rawAppConfig } = props;
  const { appID } = useParams();
  const {
    updateAppConfig,
    loading: updatingAppConfig,
    error: updateAppConfigError,
  } = useUpdateAppConfigMutation(appID);

  const initialState = useMemo(() => {
    return {
      client_id: genRandomHexadecimalString(),
      grant_types: ["authorization_code", "refresh_token"],
      response_types: ["code", "none"],
      redirect_uris: [],
      access_token_lifetime_seconds: undefined,
      refresh_token_lifetime_seconds: undefined,
      post_logout_redirect_uris: undefined,
    };
  }, []);

  const [clientConfig, setClientConfig] = useState<OAuthClientConfig>(
    initialState
  );

  const [
    createClientSuccessDialogVisible,
    setCreateClientSuccessDialogVisible,
  ] = useState(false);

  const onClientConfigChange = useCallback(
    (newClientConfig: OAuthClientConfig) => {
      setClientConfig(newClientConfig);
    },
    []
  );

  const onCreateClientSuccess = useCallback(() => {
    setCreateClientSuccessDialogVisible(true);
  }, []);

  const onCreateClick = useCallback(() => {
    // TODO: handle name field after backend is updated
    const newAppConfig = produce(rawAppConfig, (draftConfig) => {
      draftConfig.oauth = draftConfig.oauth ?? {};
      draftConfig.oauth.clients = draftConfig.oauth.clients ?? [];
      draftConfig.oauth.clients.push(clientConfig);

      clearEmptyObject(draftConfig);
    });

    updateAppConfig(newAppConfig)
      .then((result) => {
        if (result != null) {
          onCreateClientSuccess();
        }
      })
      // TODO: handle error
      .catch(() => {});
  }, [rawAppConfig, clientConfig, onCreateClientSuccess, updateAppConfig]);

  const isFormModified = useMemo(() => {
    return !deepEqual(initialState, clientConfig);
  }, [clientConfig, initialState]);

  return (
    <form className={styles.form}>
      <NavigationBlockerDialog blockNavigation={isFormModified} />
      <CreateClientSuccessDialog
        visible={createClientSuccessDialogVisible}
        clientId={clientConfig.client_id}
      />
      <ModifyOAuthClientForm
        className={styles.modifyClientForm}
        clientConfig={clientConfig}
        onClientConfigChange={onClientConfigChange}
      />
      <ButtonWithLoading
        onClick={onCreateClick}
        disabled={!isFormModified}
        labelId="create"
        loading={updatingAppConfig}
      />
    </form>
  );
};

const CreateOAuthClientScreen: React.FC = function CreateOAuthClientScreen() {
  const { appID } = useParams();
  const { data, loading, error, refetch } = useAppConfigQuery(appID);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: "..",
        label: <FormattedMessage id="OAuthClientConfiguration.title" />,
      },
      {
        to: ".",
        label: <FormattedMessage id="CreateOAuthClientScreen.title" />,
      },
    ];
  }, []);

  const { rawAppConfig, effectiveAppConfig } = useMemo(() => {
    const appConfigNode = data?.node?.__typename === "App" ? data.node : null;
    return {
      rawAppConfig: appConfigNode?.rawAppConfig ?? null,
      effectiveAppConfig: appConfigNode?.effectiveAppConfig ?? null,
    };
  }, [data]);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  if (rawAppConfig == null || effectiveAppConfig == null) {
    return null;
  }

  return (
    <main className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <CreateOAuthClientForm rawAppConfig={rawAppConfig} />
    </main>
  );
};

export default CreateOAuthClientScreen;
