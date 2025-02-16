# kubehoist

A project/example of a controller which helps spin up other CRD-based controllers, but only when needed.

Specifically, it allows you to install CRDs and map them to a controller _without_ installing the controller/application up-front.
This allows you to pre-install CRDs into a cluster which you may or may not use later, without spending the resources on running the controller(s) until they're needed (when the corresponding CRD is actually used).

Right now, this is mostly a POC which uses helm as the source of CRDs/controllers to manage. This could be expanded to other installation mechanisms as well.

## Usage

Install the controller

TODO: Add install instructions

Once the controller is installed, you can tell it to install the CRDs from another helm chart application and start monitoring them.

This is done via the `controller.kubehoist.io/v1alpha1` API `ControllerWatch` CRD which is installed as part of this controller.

Here's an example using cert manager

```yaml
apiVersion: controller.kubehoist.io/v1alpha1
kind: ControllerWatch
metadata:
  name: certmanager-sample
spec:
  helmSpec:
    chart: oci://registry-1.docker.io/bitnamicharts/cert-manager
    namespace: default
    releaseName: certmanager
    values: |
      installCRDs: true
```

Once you apply this to the cluster, you should be able to get the status. If everything is working, kubehoist should have installed the CRDs from this application automatically. You can check this on the status of the resource we just created.

```shell
$ kubectl get controllerwatch certmanager-sample -o yaml

apiVersion: controller.kubehoist.io/v1alpha1
kind: ControllerWatch
...
status:
  crdInstallationStatus: Installed
  installedCRDs:
  - group: cert-manager.io
    kind: CertificateRequest
    version: v1
  - group: cert-manager.io
    kind: Certificate
    version: v1
  - group: acme.cert-manager.io
    kind: Challenge
    version: v1
  - group: cert-manager.io
    kind: ClusterIssuer
    version: v1
  - group: cert-manager.io
    kind: Issuer
    version: v1
  - group: acme.cert-manager.io
    kind: Order
    version: v1
  lastUpdated: "2025-02-16T07:25:06Z"
```

We can also validate that the CRDs are installed directly:

```shell
$ kubectl get crd

NAME                                        CREATED AT
certificaterequests.cert-manager.io         2025-02-16T07:25:05Z
certificates.cert-manager.io                2025-02-16T07:25:05Z
challenges.acme.cert-manager.io             2025-02-16T07:25:05Z
clusterissuers.cert-manager.io              2025-02-16T07:25:05Z
controllerwatches.controller.kubehoist.io   2025-02-16T07:23:47Z
issuers.cert-manager.io                     2025-02-16T07:25:05Z
orders.acme.cert-manager.io                 2025-02-16T07:25:05Z
```

We can also check to show that the cert manager application hasn't actually been installed yet:

```shell
$ kubectl get all -n default

NAME                 TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
service/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   5m56s
```

Now try using one of the CRDs. For example, creating a self-signed cert issuer with the cert manager CRD:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: selfsigned-issuer
  namespace: default
spec:
  selfSigned: {}
```

Once applied to the cluster, kubehoist will see that the cert-manager CRD is used, and automatically install cert manager itself according to the settings provided earlier in the `ControllerWatch` resource. This may take a few moments to install.

In order to confirm it worked, you can confirm that the issuer we created is marked as ready, indicating that cert manager has successfully picked up and created this issuer (the ready column would be empty if cert manager did not successfully see this resource):

```shell
$ kubectl get issuer selfsigned-issuer

NAME                READY   AGE
selfsigned-issuer   True    32s
```

You can also see cert manager resources running now:

```shell
kubectl get all -n default

NAME                                                       READY   STATUS    RESTARTS   AGE
pod/certmanager-cert-manager-cainjector-58d46db984-xwkhm   1/1     Running   0          49s
pod/certmanager-cert-manager-controller-6b4f79c597-4bxkf   1/1     Running   0          49s
pod/certmanager-cert-manager-webhook-6f86cb86c9-q57st      1/1     Running   0          49s

NAME                                                  TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/certmanager-cert-manager-controller-metrics   ClusterIP   10.96.206.237   <none>        9402/TCP   49s
service/certmanager-cert-manager-webhook              ClusterIP   10.96.136.189   <none>        443/TCP    49s
service/kubernetes                                    ClusterIP   10.96.0.1       <none>        443/TCP    10m

NAME                                                  READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/certmanager-cert-manager-cainjector   1/1     1            1           49s
deployment.apps/certmanager-cert-manager-controller   1/1     1            1           49s
deployment.apps/certmanager-cert-manager-webhook      1/1     1            1           49s

NAME                                                             DESIRED   CURRENT   READY   AGE
replicaset.apps/certmanager-cert-manager-cainjector-58d46db984   1         1         1       49s
replicaset.apps/certmanager-cert-manager-controller-6b4f79c597   1         1         1       49s
replicaset.apps/certmanager-cert-manager-webhook-6f86cb86c9      1         1         1       49s
```

Note, the cert manager application itself won't get installed until you actual use one of the cert-manager CRDs.

## Notes

There are various kubernetes applications which contain CRDs which also install validating or mutating webhooks, watching those installed CRDs.
This goes beyond a simple controller watching the CRD itself.

For these cases, since we explicitly aren't installing any resources to kubernetes except the CRDs themselves upfront, any sort of validating/mutating webhook provided by the application may not apply for the first usage
before kubehoist has detected CRD usage and installed the actual application/spun up its components.
