/*
 * Copyright (c) 2022-2026 Super Durable, Inc.
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

package io.iworkflow.core.mapper;

import io.iworkflow.core.InternalConditionalClose;
import io.iworkflow.core.ObjectEncoder;
import io.iworkflow.core.Registry;
import io.iworkflow.gen.models.EncodedObject;
import io.iworkflow.gen.models.StateDecision;
import io.iworkflow.gen.models.WorkflowConditionalClose;

import java.util.stream.Collectors;

public class StateDecisionMapper {
    public static StateDecision toGenerated(io.iworkflow.core.StateDecision stateDecision, final String workflowType, final Registry registry, final ObjectEncoder objectEncoder) {
        if (stateDecision.getNextStates() == null && !stateDecision.getWorkflowConditionalClose().isPresent()) {
            return null;
        }

        StateDecision decision = new StateDecision();

        if (stateDecision.getNextStates() != null) {
            decision.nextStates(
                    stateDecision.getNextStates()
                            .stream()
                            .map(movement -> StateMovementMapper.toGenerated(movement, workflowType, registry, objectEncoder))
                            .collect(Collectors.toList())
            );
        }

        if (!stateDecision.getWorkflowConditionalClose().isPresent()) {
            return decision;
        }

        InternalConditionalClose conditionalClose = stateDecision.getWorkflowConditionalClose().get();
        EncodedObject closeInput = objectEncoder.encode(conditionalClose.getCloseInput());
        decision.conditionalClose(
                new WorkflowConditionalClose()
                        .conditionalCloseType(conditionalClose.getWorkflowConditionalCloseType())
                        .closeInput(closeInput)
                        .channelName(conditionalClose.getChannelName())
        );
        return decision;
    }
}
