/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tests

import client "github.com/devtron-labs/kubelink/grpc"

const clusterName = "default_cluster"
const clusterId int32 = 1

var clusterConfig = &client.ClusterConfig{
	ApiServerUrl:          "https://kubernetes.default.svc",
	Token:                 "",
	ClusterId:             clusterId,
	ClusterName:           clusterName,
	InsecureSkipTLSVerify: true,
}

const deploymentReferenceTemplateDir = "/tmp/deployment-chart_4-18-0"
const cronjobReferenceTemplateDir = "/tmp/cronjob-chart_1-5-0"
const statefulSetReferenceTemplateDir = "/tmp/statefulset-chart_5-0-0"
const rolloutReferenceTemplateDir = "/tmp/rollout-reference-chart_4-18-0"

// deployment variables
const deploymentReleaseName = "deployment-test1"
const deploymentReleaseNamespace = "devtron-demo"

// rollout variables
const rolloutReleaseName = "rollout-devtron-demo"
const rolloutReleaseNamespace = "devtron-demo"

// statefulSet variables
const statefulSetReleaseName = "statefulset-devtron-demo"
const statefulSetReleaseNamespace = "devtron-demo"

// cronjob and job variables
const jobCronjobReleaseName = "cronjob-devtron-demo"
const jobCronjobReleaseNamespace = "devtron-demo"

const appName = "sample-app"
const chartVersion = "4.18.1"
const apiVersion = "v1"

const DeploymentYamlValue = "{\"ConfigMaps\":{\"enabled\":false},\"ConfigSecrets\":{\"enabled\":false,\"secrets\":[]},\"ContainerPort\":[{\"envoyPort\":8799,\"idleTimeout\":\"1800s\",\"name\":\"app\",\"port\":8080,\"servicePort\":80,\"supportStreaming\":false,\"useHTTP2\":false}],\"EnvVariables\":[],\"GracePeriod\":30,\"LivenessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"MaxSurge\":1,\"MaxUnavailable\":0,\"MinReadySeconds\":60,\"ReadinessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"Spec\":{\"Affinity\":{\"Key\":\"\",\"Values\":\"\"}},\"StartupProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"ambassadorMapping\":{\"ambassadorId\":\"\",\"cors\":{},\"enabled\":false,\"hostname\":\"devtron.example.com\",\"labels\":{},\"prefix\":\"/\",\"retryPolicy\":{},\"rewrite\":\"\",\"tls\":{\"context\":\"\",\"create\":false,\"hosts\":[],\"secretName\":\"\"}},\"app\":\"287\",\"appLabels\":{},\"appMetrics\":false,\"args\":{\"enabled\":false,\"value\":[\"/bin/sh\",\"-c\",\"touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600\"]},\"autoscaling\":{\"MaxReplicas\":2,\"MinReplicas\":1,\"TargetCPUUtilizationPercentage\":90,\"TargetMemoryUtilizationPercentage\":80,\"annotations\":{},\"behavior\":{},\"enabled\":false,\"extraMetrics\":[],\"labels\":{}},\"command\":{\"enabled\":false,\"value\":[],\"workingDir\":{}},\"containerSecurityContext\":{},\"containerSpec\":{\"lifecycle\":{\"enabled\":false,\"postStart\":{\"httpGet\":{\"host\":\"example.com\",\"path\":\"/example\",\"port\":90}},\"preStop\":{\"exec\":{\"command\":[\"sleep\",\"10\"]}}}},\"containers\":[],\"dbMigrationConfig\":{\"enabled\":false},\"deployment\":{\"strategy\":{\"rolling\":{\"maxSurge\":\"25%\",\"maxUnavailable\":1}}},\"deploymentAnnotations\":{},\"deploymentLabels\":{},\"deploymentType\":\"ROLLING\",\"env\":\"1\",\"envoyproxy\":{\"configMapName\":\"\",\"image\":\"docker.io/envoyproxy/envoy:v1.16.0\",\"lifecycle\":{},\"resources\":{\"limits\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"}}},\"flaggerCanary\":{\"addOtherGateways\":[],\"addOtherHosts\":[],\"analysis\":{\"interval\":\"15s\",\"maxWeight\":50,\"stepWeight\":5,\"threshold\":5},\"annotations\":{},\"appProtocol\":\"http\",\"createIstioGateway\":{\"annotations\":{},\"enabled\":false,\"labels\":{},\"tls\":{\"enabled\":false}},\"enabled\":false,\"labels\":{},\"loadtest\":{\"enabled\":true,\"url\":\"http://flagger-loadtester.istio-system/\"},\"match\":[{\"uri\":{\"prefix\":\"/\"}}],\"portDiscovery\":true,\"rewriteUri\":\"/\",\"serviceport\":8080,\"targetPort\":8080,\"thresholds\":{\"latency\":500,\"successRate\":90}},\"hostAliases\":[],\"image\":{\"pullPolicy\":\"IfNotPresent\"},\"imagePullSecrets\":[],\"ingress\":{\"annotations\":{},\"className\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.local\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example1\"]}],\"labels\":{},\"tls\":[]},\"ingressInternal\":{\"annotations\":{},\"className\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example1\"]},{\"host\":\"chart-example2.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example2\",\"/example2/healthz\"]}],\"tls\":[]},\"initContainers\":[],\"istio\":{\"authorizationPolicy\":{\"annotations\":{},\"enabled\":false,\"labels\":{},\"provider\":{},\"rules\":[]},\"destinationRule\":{\"annotations\":{},\"enabled\":false,\"labels\":{},\"subsets\":[],\"trafficPolicy\":{}},\"enable\":false,\"gateway\":{\"annotations\":{},\"enabled\":false,\"host\":\"example.com\",\"labels\":{},\"tls\":{\"enabled\":false,\"secretName\":\"example-secret\"}},\"peerAuthentication\":{\"annotations\":{},\"enabled\":false,\"labels\":{},\"mtls\":{\"mode\":\"\"},\"portLevelMtls\":{},\"selector\":{\"enabled\":false}},\"requestAuthentication\":{\"annotations\":{},\"enabled\":false,\"jwtRules\":[],\"labels\":{},\"selector\":{\"enabled\":false}},\"virtualService\":{\"annotations\":{},\"enabled\":false,\"gateways\":[],\"hosts\":[],\"http\":[],\"labels\":{}}},\"kedaAutoscaling\":{\"advanced\":{},\"authenticationRef\":{},\"enabled\":false,\"envSourceContainerName\":\"\",\"maxReplicaCount\":2,\"minReplicaCount\":1,\"triggerAuthentication\":{\"enabled\":false,\"name\":\"\",\"spec\":{}},\"triggers\":[]},\"networkPolicy\":{\"annotations\":{},\"egress\":[],\"enabled\":false,\"ingress\":[],\"labels\":{},\"podSelector\":{\"matchExpressions\":[],\"matchLabels\":{}},\"policyTypes\":[]},\"pauseForSecondsBeforeSwitchActive\":30,\"pipelineName\":\"cd-287-pp0e\",\"podAnnotations\":{},\"podDisruptionBudget\":{},\"podLabels\":{},\"podSecurityContext\":{},\"prometheus\":{\"release\":\"monitoring\"},\"rawYaml\":[],\"releaseVersion\":\"2\",\"replicaCount\":1,\"resources\":{\"limits\":{\"cpu\":\"0.05\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"0.01\",\"memory\":\"10Mi\"}},\"restartPolicy\":\"Always\",\"secret\":{\"data\":{},\"enabled\":false},\"server\":{\"deployment\":{\"image\":\"rish2320/rishabhapp-1\",\"image_tag\":\"6a824121-150-33\"}},\"service\":{\"annotations\":{},\"loadBalancerSourceRanges\":[],\"type\":\"ClusterIP\"},\"serviceAccount\":{\"annotations\":{},\"create\":false,\"name\":\"\"},\"servicemonitor\":{\"additionalLabels\":{}},\"tolerations\":[],\"topologySpreadConstraints\":[],\"volumeMounts\":[],\"volumes\":[],\"waitForSecondsBeforeScalingDown\":30,\"winterSoldier\":{\"action\":\"sleep\",\"annotation\":{},\"apiVersion\":\"pincher.devtron.ai/v1alpha1\",\"enabled\":false,\"fieldSelector\":[\"AfterTime(AddTime(ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '5m'), Now())\"],\"labels\":{},\"targetReplicas\":[],\"timeRangesWithZone\":{\"timeRanges\":[],\"timeZone\":\"Asia/Kolkata\"},\"type\":\"Deployment\"}}"
const CronJobYamlValue = "{\"ConfigMaps\":{\"enabled\":false},\"ConfigSecrets\":{\"enabled\":false,\"secrets\":[]},\"ContainerPort\":[{\"envoyPort\":8799,\"idleTimeout\":\"1800s\",\"name\":\"app\",\"port\":8080,\"servicePort\":80,\"supportStreaming\":true,\"useHTTP2\":true}],\"EnvVariables\":[],\"GracePeriod\":30,\"MaxSurge\":1,\"MaxUnavailable\":0,\"MinReadySeconds\":60,\"Spec\":{\"Affinity\":{\"Values\":\"nodes\",\"key\":\"\"}},\"app\":\"282\",\"appLabels\":{},\"appMetrics\":false,\"args\":{\"enabled\":false,\"value\":[\"/bin/sh\",\"-c\",\"touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600\"]},\"command\":{\"enabled\":false,\"value\":[]},\"containerSecurityContext\":{\"allowPrivilegeEscalation\":false},\"containers\":[],\"cronjobConfigs\":{\"concurrencyPolicy\":\"Allow\",\"failedJobsHistoryLimit\":1,\"restartPolicy\":\"OnFailure\",\"schedule\":\"* * * * *\",\"startingDeadlineSeconds\":100,\"successfulJobsHistoryLimit\":3,\"suspend\":false},\"dbMigrationConfig\":{\"enabled\":false},\"deployment\":{\"strategy\":{\"rolling\":{\"maxSurge\":\"25%\",\"maxUnavailable\":1}}},\"deploymentType\":\"ROLLING\",\"env\":\"1\",\"ephemeralContainers\":[],\"image\":{\"pullPolicy\":\"IfNotPresent\"},\"imagePullSecrets\":[],\"initContainers\":[],\"jobConfigs\":{\"activeDeadlineSeconds\":100,\"backoffLimit\":5,\"completions\":2,\"parallelism\":1,\"suspend\":false},\"kedaAutoscaling\":{\"envSourceContainerName\":\"\",\"failedJobsHistoryLimit\":5,\"maxReplicaCount\":2,\"minReplicaCount\":1,\"pollingInterval\":30,\"rolloutStrategy\":\"default\",\"scalingStrategy\":{\"customScalingQueueLengthDeduction\":1,\"customScalingRunningJobPercentage\":\"0.5\",\"multipleScalersCalculation\":\"max\",\"pendingPodConditions\":[\"Ready\",\"PodScheduled\",\"AnyOtherCustomPodCondition\"],\"strategy\":\"custom\"},\"successfulJobsHistoryLimit\":5,\"triggerAuthentication\":{\"enabled\":false,\"name\":\"\",\"spec\":{}},\"triggers\":[{\"authenticationRef\":{},\"metadata\":{\"host\":\"RabbitMqHost\",\"queueLength\":\"5\",\"queueName\":\"hello\"},\"type\":\"rabbitmq\"}]},\"kind\":\"Job\",\"pauseForSecondsBeforeSwitchActive\":30,\"pipelineName\":\"cd-282-odk0\",\"podAnnotations\":{},\"podLabels\":{},\"podSecurityContext\":{},\"prometheus\":{\"release\":\"monitoring\"},\"rawYaml\":[],\"readinessGates\":[],\"releaseVersion\":\"7\",\"resources\":{\"limits\":{\"cpu\":\"0.05\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"0.01\",\"memory\":\"10Mi\"}},\"secret\":{\"data\":{},\"enabled\":false},\"server\":{\"deployment\":{\"image\":\"rish2320/rishabhapp-1\",\"image_tag\":\"6a824121-146-29\"}},\"service\":{\"annotations\":{},\"enabled\":false,\"type\":\"ClusterIP\"},\"servicemonitor\":{\"additionalLabels\":{}},\"setHostnameAsFQDN\":false,\"shareProcessNamespace\":false,\"tolerations\":[],\"topologySpreadConstraints\":[],\"volumeMounts\":[],\"volumes\":[],\"waitForSecondsBeforeScalingDown\":30}"
const RollOutYamlValue = "{\"ConfigMaps\":{\"enabled\":false},\"ConfigSecrets\":{\"enabled\":false,\"secrets\":[]},\"ContainerPort\":[{\"envoyPort\":8799,\"idleTimeout\":\"1800s\",\"name\":\"app\",\"port\":8080,\"servicePort\":80,\"supportStreaming\":false,\"useHTTP2\":false}],\"EnvVariables\":[],\"GracePeriod\":30,\"LivenessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"MaxSurge\":1,\"MaxUnavailable\":0,\"MinReadySeconds\":60,\"ReadinessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"Spec\":{\"Affinity\":{\"Values\":\"nodes\",\"key\":\"\"}},\"StartupProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"ambassadorMapping\":{\"ambassadorId\":\"\",\"cors\":{},\"enabled\":false,\"hostname\":\"devtron.example.com\",\"labels\":{},\"prefix\":\"/\",\"retryPolicy\":{},\"rewrite\":\"\",\"tls\":{\"context\":\"\",\"create\":false,\"hosts\":[],\"secretName\":\"\"}},\"app\":\"292\",\"appLabels\":{},\"appMetrics\":false,\"args\":{\"enabled\":false,\"value\":[\"/bin/sh\",\"-c\",\"touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600\"]},\"autoscaling\":{\"MaxReplicas\":2,\"MinReplicas\":1,\"TargetCPUUtilizationPercentage\":90,\"TargetMemoryUtilizationPercentage\":80,\"annotations\":{},\"behavior\":{},\"enabled\":false,\"extraMetrics\":[],\"labels\":{}},\"command\":{\"enabled\":false,\"value\":[],\"workingDir\":{}},\"containerSecurityContext\":{},\"containerSpec\":{\"lifecycle\":{\"enabled\":false,\"postStart\":{\"httpGet\":{\"host\":\"example.com\",\"path\":\"/example\",\"port\":90}},\"preStop\":{\"exec\":{\"command\":[\"sleep\",\"10\"]}}}},\"containers\":[],\"dbMigrationConfig\":{\"enabled\":false},\"deployment\":{\"strategy\":{\"rolling\":{\"maxSurge\":\"25%\",\"maxUnavailable\":1}}},\"deploymentType\":\"ROLLING\",\"env\":\"1\",\"envoyproxy\":{\"configMapName\":\"\",\"image\":\"docker.io/envoyproxy/envoy:v1.16.0\",\"lifecycle\":{},\"resources\":{\"limits\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"}}},\"hostAliases\":[],\"image\":{\"pullPolicy\":\"IfNotPresent\"},\"imagePullSecrets\":[],\"ingress\":{\"annotations\":{},\"className\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.local\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example1\"]}],\"labels\":{},\"tls\":[]},\"ingressInternal\":{\"annotations\":{},\"className\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example1\"]},{\"host\":\"chart-example2.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example2\",\"/example2/healthz\"]}],\"tls\":[]},\"initContainers\":[],\"istio\":{\"authorizationPolicy\":{\"annotations\":{},\"enabled\":false,\"labels\":{},\"provider\":{},\"rules\":[]},\"destinationRule\":{\"annotations\":{},\"enabled\":false,\"labels\":{},\"subsets\":[],\"trafficPolicy\":{}},\"enable\":false,\"gateway\":{\"annotations\":{},\"enabled\":false,\"host\":\"example.com\",\"labels\":{},\"tls\":{\"enabled\":false,\"secretName\":\"secret-name\"}},\"peerAuthentication\":{\"annotations\":{},\"enabled\":false,\"labels\":{},\"mtls\":{\"mode\":\"\"},\"portLevelMtls\":{},\"selector\":{\"enabled\":false}},\"requestAuthentication\":{\"annotations\":{},\"enabled\":false,\"jwtRules\":[],\"labels\":{},\"selector\":{\"enabled\":false}},\"virtualService\":{\"annotations\":{},\"enabled\":false,\"gateways\":[],\"hosts\":[],\"http\":[],\"labels\":{}}},\"kedaAutoscaling\":{\"advanced\":{},\"authenticationRef\":{},\"enabled\":false,\"envSourceContainerName\":\"\",\"maxReplicaCount\":2,\"minReplicaCount\":1,\"triggerAuthentication\":{\"enabled\":false,\"name\":\"\",\"spec\":{}},\"triggers\":[]},\"networkPolicy\":{\"annotations\":{},\"egress\":[],\"enabled\":false,\"ingress\":[],\"labels\":{},\"podSelector\":{\"matchExpressions\":[],\"matchLabels\":{}},\"policyTypes\":[]},\"pauseForSecondsBeforeSwitchActive\":30,\"pipelineName\":\"cd-292-1jwq\",\"podAnnotations\":{},\"podDisruptionBudget\":{},\"podLabels\":{},\"podSecurityContext\":{},\"prometheus\":{\"release\":\"monitoring\"},\"rawYaml\":[],\"releaseVersion\":\"2\",\"replicaCount\":1,\"resources\":{\"limits\":{\"cpu\":\"0.05\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"0.01\",\"memory\":\"10Mi\"}},\"restartPolicy\":\"Always\",\"rolloutAnnotations\":{},\"rolloutLabels\":{},\"secret\":{\"data\":{},\"enabled\":false},\"server\":{\"deployment\":{\"image\":\"rish2320/rishabhapp-1\",\"image_tag\":\"6a824121-152-35\"}},\"service\":{\"annotations\":{},\"loadBalancerSourceRanges\":[],\"type\":\"ClusterIP\"},\"serviceAccount\":{\"annotations\":{},\"create\":false,\"name\":\"\"},\"servicemonitor\":{\"additionalLabels\":{}},\"tolerations\":[],\"topologySpreadConstraints\":[],\"volumeMounts\":[],\"volumes\":[],\"waitForSecondsBeforeScalingDown\":30,\"winterSoldier\":{\"action\":\"sleep\",\"annotation\":{},\"apiVersion\":\"pincher.devtron.ai/v1alpha1\",\"enabled\":false,\"fieldSelector\":[\"AfterTime(AddTime(ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '5m'), Now())\"],\"labels\":{},\"targetReplicas\":[],\"timeRangesWithZone\":{\"timeRanges\":[],\"timeZone\":\"Asia/Kolkata\"},\"type\":\"Rollout\"}}"
const StatefulSetYamlValue = "{\"ConfigMaps\":{\"enabled\":false},\"ConfigSecrets\":{\"enabled\":false,\"secrets\":[]},\"ContainerPort\":[{\"envoyPort\":8799,\"idleTimeout\":\"1800s\",\"name\":\"app\",\"port\":8080,\"servicePort\":80,\"supportStreaming\":false,\"useHTTP2\":false}],\"EnvVariables\":[],\"EnvVariablesFromCongigMapKeys\":[],\"EnvVariablesFromSecretKeys\":[],\"GracePeriod\":30,\"LivenessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"MaxSurge\":1,\"MaxUnavailable\":0,\"MinReadySeconds\":60,\"ReadinessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"Spec\":{\"Affinity\":{\"Values\":\"nodes\",\"key\":\"\"}},\"ambassadorMapping\":{\"ambassadorId\":\"\",\"cors\":{},\"enabled\":false,\"hostname\":\"devtron.example.com\",\"labels\":{},\"prefix\":\"/\",\"retryPolicy\":{},\"rewrite\":\"\",\"tls\":{\"context\":\"\",\"create\":false,\"hosts\":[],\"secretName\":\"\"}},\"app\":\"2\",\"appLabels\":{},\"appMetrics\":false,\"args\":{\"enabled\":false,\"value\":[\"/bin/sh\",\"-c\",\"touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600\"]},\"autoscaling\":{\"MaxReplicas\":2,\"MinReplicas\":1,\"TargetCPUUtilizationPercentage\":90,\"TargetMemoryUtilizationPercentage\":80,\"annotations\":{},\"behavior\":{},\"enabled\":false,\"extraMetrics\":[],\"labels\":{}},\"command\":{\"enabled\":false,\"value\":[],\"workingDir\":{}},\"containerSecurityContext\":{},\"containerSpec\":{\"lifecycle\":{\"enabled\":false,\"postStart\":{\"httpGet\":{\"host\":\"example.com\",\"path\":\"/example\",\"port\":90}},\"preStop\":{\"exec\":{\"command\":[\"sleep\",\"10\"]}}}},\"containers\":[],\"dbMigrationConfig\":{\"enabled\":false},\"deployment\":{\"strategy\":{\"rollingUpdate\":{\"partition\":0}}},\"deploymentType\":\"ROLLINGUPDATE\",\"env\":\"1\",\"envoyproxy\":{\"configMapName\":\"\",\"image\":\"quay.io/devtron/envoy:v1.14.1\",\"lifecycle\":{},\"resources\":{\"limits\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"}}},\"hostAliases\":[],\"image\":{\"pullPolicy\":\"IfNotPresent\"},\"imagePullSecrets\":[],\"ingress\":{\"annotations\":{},\"className\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.local\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example1\"]},{\"host\":\"chart-example2.local\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example2\",\"/example2/healthz\"]}],\"labels\":{},\"tls\":[]},\"ingressInternal\":{\"annotations\":{},\"className\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example1\"]},{\"host\":\"chart-example2.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example2\",\"/example2/healthz\"]}],\"tls\":[]},\"initContainers\":[],\"istio\":{\"enable\":false,\"gateway\":{\"annotations\":{},\"enabled\":false,\"host\":\"example.com\",\"labels\":{},\"tls\":{\"enabled\":false,\"secretName\":\"secret-name\"}},\"virtualService\":{\"annotations\":{},\"enabled\":false,\"gateways\":[],\"hosts\":[],\"http\":[{\"corsPolicy\":{},\"headers\":{},\"match\":[{\"uri\":{\"prefix\":\"/v1\"}},{\"uri\":{\"prefix\":\"/v2\"}}],\"retries\":{\"attempts\":2,\"perTryTimeout\":\"3s\"},\"rewriteUri\":\"/\",\"route\":[{\"destination\":{\"host\":\"service1\",\"port\":80}}],\"timeout\":\"12s\"},{\"route\":[{\"destination\":{\"host\":\"service2\"}}]}],\"labels\":{}}},\"kedaAutoscaling\":{\"advanced\":{},\"authenticationRef\":{},\"enabled\":false,\"envSourceContainerName\":\"\",\"maxReplicaCount\":2,\"minReplicaCount\":1,\"triggerAuthentication\":{\"enabled\":false,\"name\":\"\",\"spec\":{}},\"triggers\":[]},\"pauseForSecondsBeforeSwitchActive\":30,\"pipelineName\":\"cd-2-mxd6\",\"podAnnotations\":{},\"podLabels\":{},\"podSecurityContext\":{},\"prometheus\":{\"release\":\"monitoring\"},\"rawYaml\":[],\"releaseVersion\":\"4\",\"replicaCount\":1,\"resources\":{\"limits\":{\"cpu\":\"0.05\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"0.01\",\"memory\":\"10Mi\"}},\"secret\":{\"data\":{},\"enabled\":false},\"server\":{\"deployment\":{\"image\":\"rish2320/rishabhapp-1\",\"image_tag\":\"6a824121-1-1\"}},\"service\":{\"annotations\":{},\"enabled\":false,\"loadBalancerSourceRanges\":[],\"type\":\"ClusterIP\"},\"serviceAccount\":{\"annotations\":{},\"create\":false,\"name\":\"\"},\"servicemonitor\":{\"additionalLabels\":{}},\"statefulSetConfig\":{\"annotations\":{},\"labels\":{},\"volumeClaimTemplates\":[]},\"tolerations\":[],\"topologySpreadConstraints\":[],\"volumeMounts\":[],\"volumes\":[{\"accessModes\":[\"ReadWriteOnce\"],\"resources\":{\"requests\":{\"storage\":\"2Gi\"}}}],\"waitForSecondsBeforeScalingDown\":30,\"winterSoldier\":{\"action\":\"sleep\",\"annotation\":{},\"apiVersion\":\"pincher.devtron.ai/v1alpha1\",\"enabled\":false,\"fieldSelector\":[],\"labels\":{},\"targetReplicas\":[],\"timeRangesWithZone\":{\"timeRanges\":[],\"timeZone\":\"Asia/Kolkata\"}}}"
const InstallReleaseReqYamlValue = "## Reference to one or more secrets to be used when pulling images\n## ref: https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/\nimagePullSecrets: []\n# - name: \"image-pull-secret\"\n## Operator\noperator:\n  # Name that will be assigned to most of internal Kubernetes objects like\n  # Deployment, ServiceAccount, Role etc.\n  name: mongodb-kubernetes-operator\n\n  # Name of the operator image\n  operatorImageName: mongodb-kubernetes-operator\n\n  # Name of the deployment of the operator pod\n  deploymentName: mongodb-kubernetes-operator\n\n  # Version of mongodb-kubernetes-operator\n  version: 0.8.3\n\n  # Uncomment this line to watch all namespaces\n  # watchNamespace: \"*\"\n\n  # Resources allocated to Operator Pod\n  resources:\n    limits:\n      cpu: 1100m\n      memory: 1Gi\n    requests:\n      cpu: 500m\n      memory: 200Mi\n\n  # replicas deployed for the operator pod. Running 1 is optimal and suggested.\n  replicas: 1\n\n  # Additional environment variables\n  extraEnvs: []\n  # environment:\n  # - name: CLUSTER_DOMAIN\n  #   value: my-cluster.domain\n\n  podSecurityContext:\n    runAsNonRoot: true\n    runAsUser: 2000\n\n  securityContext: {}\n\n## Operator's database\ndatabase:\n  name: mongodb-database\n  # set this to the namespace where you would like\n  # to deploy the MongoDB database,\n  # Note if the database namespace is not same\n  # as the operator namespace,\n  # make sure to set \"watchNamespace\" to \"*\"\n  # to ensure that the operator has the\n  # permission to reconcile resources in other namespaces\n  # namespace: mongodb-database\n\nagent:\n  name: mongodb-agent\n  version: 12.0.25.7724-1\nversionUpgradeHook:\n  name: mongodb-kubernetes-operator-version-upgrade-post-start-hook\n  version: 1.0.8\nreadinessProbe:\n  name: mongodb-kubernetes-readinessprobe\n  version: 1.0.17\nmongodb:\n  name: mongo\n  repo: docker.io\n\nregistry:\n  agent: quay.io/mongodb\n  versionUpgradeHook: quay.io/mongodb\n  readinessProbe: quay.io/mongodb\n  operator: quay.io/mongodb\n  pullPolicy: Always\n\n# Set to false if CRDs have been installed already. The CRDs can be installed\n# manually from the code repo: github.com/mongodb/mongodb-kubernetes-operator or\n# using the `community-operator-crds` Helm chart.\ncommunity-operator-crds:\n  enabled: true\n\n# Deploys MongoDB with `resource` attributes.\ncreateResource: false\nresource:\n  name: mongodb-replica-set\n  version: 4.4.0\n  members: 3\n  tls:\n    enabled: false\n\n    # Installs Cert-Manager in this cluster.\n    useX509: false\n    sampleX509User: false\n    useCertManager: true\n    certificateKeySecretRef: tls-certificate\n    caCertificateSecretRef: tls-ca-key-pair\n    certManager:\n      certDuration: 8760h   # 365 days\n      renewCertBefore: 720h   # 30 days\n\n  users: []\n  # if using the MongoDBCommunity Resource, list any users to be added to the resource\n  # users:\n  # - name: my-user\n  #   db: admin\n  #   passwordSecretRef: # a reference to the secret that will be used to generate the user's password\n  #     name: <secretName>\n  #   roles:\n  #     - name: clusterAdmin\n  #       db: admin\n  #     - name: userAdminAnyDatabase\n  #       db: admin\n  #     - name: readWriteAnyDatabase\n  #       db: admin\n  #     - name: dbAdminAnyDatabase\n  #       db: admin\n  #   scramCredentialsSecretName: my-scram\n"
