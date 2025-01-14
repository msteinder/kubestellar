<!--example1-kcp-provider-start-->

#### Create a space provider description for KCP

Space provider for KCP will allow you to use KCP as backend provider for spaces.
Use the following commands to create a provider secret for KCP access and
a space provider definition.

```shell
KUBECONFIG=$SM_CONFIG kubectl create secret generic kcpsec --from-file=kubeconfig=$PROVIDER_KUBECONFIG --from-file=incluster=$PROVIDER_KUBECONFIG
KUBECONFIG=$SM_CONFIG kubectl apply -f - <<EOF
apiVersion: space.kubestellar.io/v1alpha1
kind: SpaceProviderDesc
metadata:
  name: default
spec:
  ProviderType: "kcp"
  SpacePrefixForDiscovery: "ks-"
  secretRef:
    namespace: default
    name: kcpsec
EOF
```

Next, use the following command to wait for the space-manger to process the provider.

```shell
KUBECONFIG=$SM_CONFIG kubectl wait --for=jsonpath='{.status.Phase}'=Ready spaceproviderdesc/default --timeout=90s
```

The following variable will be used in later commands to indicate that
they are being invoked close enough to the provider's apiserver to
use the more efficient networking (see [doc on
"in-cluster"](../../commands/#in-cluster)).

```shell
in_cluster="--in-cluster"
```

<!--example1-kcp-provider-end-->
