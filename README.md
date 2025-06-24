# Steadybit extension-splunk-platform

A [Steadybit](https://www.steadybit.com/) extension to integrate [Splunk Cloud Platform](https://www.splunk.com/en_us/products/splunk-cloud-platform.html)
and [Splunk Enterprise](https://www.splunk.com/en_us/products/splunk-enterprise.html).

Learn about the capabilities of this extension in our [Reliability Hub](https://hub.steadybit.com/extension/com.steadybit.extension_splunk-platform).

## Prerequisites

The extension supports Splunk token based authentication. Please follow the instructions on the Splunk documentation page
[Set up authentication with token](https://docs.splunk.com/Documentation/Splunk/latest/Security/Setupauthenticationwithtokens).

You might need to take extra steps to access your Splunk Cloud Platform deployment using the Splunk REST API. Details are available at
[You might need to take extra steps to access your Splunk Cloud Platform deployment using the Splunk REST API](https://docs.splunk.com/Documentation/Splunk/latest/RESTTUT/RESTandCloud).

Supported Splunk Cloud Platform and Splunk Enterprise versions:
- 9.4.2+

## Configuration

| Environment Variable                                      | Helm value                  | Meaning                                                                                                                              | Required | Default |
|-----------------------------------------------------------|-----------------------------|--------------------------------------------------------------------------------------------------------------------------------------|----------|---------|
| `STEADYBIT_EXTENSION_ACCESS_TOKEN`                        | `splunk.accessToken`        | The token required to access the Splunk Cloud Platform or Splunk Enterprise.                                                         | Yes      |         |
| `STEADYBIT_EXTENSION_API_BASE_URL`                        | `splunk.apiBaseUrl`         | The API URL of the Splunk Cloud Platform or Splunk Enterprise instance, for example `https://<deployment-name>.splunkcloud.com:8089` | Yes      |         |
| `STEADYBIT_EXTENSION_INSECURE_SKIP_VERIFY`                | `splunk.insecureSkipVerify` | Disable TLS certificate validation.                                                                                                  | No       | False   |
| `STEADYBIT_EXTENSION_DISCOVERY_ATTRIBUTES_EXCLUDES_ALERT` |                             | List of Alert Attributes which will be excluded during discovery. Checked by key equality and supporting trailing "*"                | No       |         |

The extension supports all environment variables provided by [steadybit/extension-kit](https://github.com/steadybit/extension-kit#environment-variables).

## Installation

### Kubernetes

Detailed information about agent and extension installation in kubernetes can also be found in
our [documentation](https://docs.steadybit.com/install-and-configure/install-agent/install-on-kubernetes).

#### Recommended (via agent helm chart)

All extensions provide a helm chart that is also integrated in the
[helm-chart](https://github.com/steadybit/helm-charts/tree/main/charts/steadybit-agent) of the agent.

You must provide additional values to activate this extension.

```
--set extension-splunk-platform.enabled=true \
```

Additional configuration options can be found in
the [helm-chart](https://github.com/steadybit/extension-splunk-platform/blob/main/charts/steadybit-extension-splunk-platform/values.yaml)
of the extension.

#### Alternative (via own helm chart)

If you need more control, you can install the extension via its
dedicated [helm-chart](https://github.com/steadybit/extension-splunk-platform/blob/main/charts/steadybit-extension-splunk-platform).

```bash
helm repo add steadybit-extension-splunk-platform https://steadybit.github.io/extension-splunk-platform
helm repo update
helm upgrade steadybit-extension-splunk-platform \
    --install \
    --wait \
    --timeout 5m0s \
    --create-namespace \
    --namespace steadybit-agent \
    steadybit-extension-splunk-platform/steadybit-extension-splunk-platform
```

### Linux Package

Please use our [agent-linux.sh script](https://docs.steadybit.com/install-and-configure/install-agent/install-on-linux-hosts)
to install the extension on your Linux machine. The script will download the latest version of the extension and install
it using the package manager.

After installing, configure the extension by editing `/etc/steadybit/extension-splunk-platform` and then restart the service.

## Extension registration

Make sure that the extension is registered with the agent. In most cases this is done automatically. Please refer to
the [documentation](https://docs.steadybit.com/install-and-configure/install-agent/extension-registration) for more
information about extension registration and how to verify.


## Importing your own certificates

You may want to import your own certificates for connecting to Jenkins instances with self-signed certificates. This can be done in two ways:

### Option 1: Using InsecureSkipVerify

The extension provides the `insecureSkipVerify` option which disables TLS certificate verification. This is suitable for testing but not recommended for production environments.

```yaml
splunk:
  insecureSkipVerify: true
```

### Option 2: Mounting custom certificates

Mount a volume with your custom certificates and reference it in `extraVolumeMounts` and `extraVolumes` in the helm chart.

This example uses a config map to store the `*.crt`-files:

```shell
kubectl create configmap -n steadybit-agent splunk-self-signed-ca --from-file=./self-signed-ca.crt
```

```yaml
extraVolumeMounts:
  - name: extra-certs
    mountPath: /etc/ssl/extra-certs
    readOnly: true
extraVolumes:
  - name: extra-certs
    configMap:
      name: splunk-self-signed-ca
extraEnv:
  - name: SSL_CERT_DIR
    value: /etc/ssl/extra-certs:/etc/ssl/certs
```


## Version and Revision

The version and revision of the extension:

- are printed during the startup of the extension
- are added as a Docker label to the image
- are available via the `version.txt`/`revision.txt` files in the root of the image
