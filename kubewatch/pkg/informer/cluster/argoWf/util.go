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

package argoWf

import (
	"fmt"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	informerBean "github.com/devtron-labs/kubewatch/pkg/informer/bean"
)

func GetNatsTopicForWorkflow(workflowType string) (string, error) {
	switch workflowType {
	case informerBean.CD_WORKFLOW_NAME:
		return pubsub.CD_WORKFLOW_STATUS_UPDATE, nil
	case informerBean.CI_WORKFLOW_NAME:
		return pubsub.WORKFLOW_STATUS_UPDATE_TOPIC, nil
	}
	return "", fmt.Errorf("no topic mapped to workflow type %s", workflowType)
}
