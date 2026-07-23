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

package io.iworkflow.integ.rpc;

import io.iworkflow.core.Context;
import io.iworkflow.core.StateDecision;
import io.iworkflow.core.WorkflowState;
import io.iworkflow.core.command.CommandRequest;
import io.iworkflow.core.command.CommandResults;
import io.iworkflow.core.communication.Communication;
import io.iworkflow.core.communication.InternalChannelCommand;
import io.iworkflow.core.persistence.Persistence;

import static io.iworkflow.integ.rpc.RpcLockingWorkflow.RPC_INTERNAL_CHANNEL_NAME;

public class RpcLockingWorkflowState1 implements WorkflowState<Void> {
    @Override
    public Class<Void> getInputType() {
        return Void.class;
    }

    @Override
    public CommandRequest waitUntil(
            Context context,
            Void input,
            Persistence persistence,
            final Communication communication) {
        return CommandRequest.forAnyCommandCompleted(
                InternalChannelCommand.create(RPC_INTERNAL_CHANNEL_NAME)
        );
    }

    @Override
    public StateDecision execute(
            Context context,
            Void input,
            CommandResults commandResults,
            Persistence persistence,
            final Communication communication) {
        return StateDecision.singleNextState(RpcLockingWorkflowState2.class);
    }
}
