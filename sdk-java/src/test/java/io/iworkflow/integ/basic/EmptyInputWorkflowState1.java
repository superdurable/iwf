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

package io.iworkflow.integ.basic;

import io.iworkflow.core.Context;
import io.iworkflow.core.StateDecision;
import io.iworkflow.core.WorkflowState;
import io.iworkflow.core.command.CommandRequest;
import io.iworkflow.core.command.CommandResults;
import io.iworkflow.core.communication.Communication;
import io.iworkflow.core.persistence.Persistence;

public class EmptyInputWorkflowState1 implements WorkflowState<Void> {

    @Override
    public Class<Void> getInputType() {
        return Void.class;
    }

    @Override
    public CommandRequest waitUntil(final Context context, final Void input, Persistence persistence, final Communication communication) {
        return CommandRequest.empty;
    }

    @Override
    public StateDecision execute(final Context context, final Void input, final CommandResults commandResults, Persistence persistence, final Communication communication) {
        return StateDecision.singleNextState(EmptyInputWorkflowState2.StateId);
    }
}