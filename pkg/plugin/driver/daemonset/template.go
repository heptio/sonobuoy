package daemonset

import (
	"github.com/heptio/sonobuoy/pkg/templates"
)

var daemonSetTemplate = templates.NewTemplate("daemonTemplate", `
---
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  annotations:
    sonobuoy-driver: DaemonSet
    sonobuoy-plugin: {{.PluginName}}
    sonobuoy-result-type: {{.ResultType}}
  labels:
    component: sonobuoy
    sonobuoy-run: '{{.SessionID}}'
    tier: analysis
  name: sonobuoy-{{.PluginName}}-daemon-set-{{.SessionID}}
  namespace: '{{.Namespace}}'
spec:
  selector:
    matchLabels:
      sonobuoy-run: '{{.SessionID}}'
  template:
    metadata:
      labels:
        component: sonobuoy
        sonobuoy-run: '{{.SessionID}}'
        tier: analysis
    spec:
      containers:
      - {{.ProducerContainer | indent 8}}
      - command: ["/run_single_node_worker.sh"]
        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: RESULTS_DIR
          value: /tmp/results
        - name: MASTER_URL
          value: '{{.MasterAddress}}'
        - name: RESULT_TYPE
          value: {{.ResultType}}
        - name: CA_CERT
          value: |
            {{.CACert | indent 12}}
        - name: CLIENT_CERT
          valueFrom:
            secretKeyRef:
              name: {{.SecretName}}
              key: tls.crt

        - name: CLIENT_KEY
          valueFrom:
            secretKeyRef:
              name: {{.SecretName}}
              key: tls.key
        image: {{.SonobuoyImage}}
        imagePullPolicy: Always
        name: sonobuoy-worker
        volumeMounts:
        - mountPath: /tmp/results
          name: results
          readOnly: false
      dnsPolicy: ClusterFirstWithHostNet
      hostIPC: true
      hostNetwork: true
      hostPID: true
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
        operator: Exists
      - key: CriticalAddonsOnly
        operator: Exists
      volumes:
      - emptyDir: {}
        name: results
      - hostPath:
          path: /
        name: root
`)
