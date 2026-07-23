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
import io.iworkflow.core.WorkflowStateOptions;
import io.iworkflow.core.command.CommandRequest;
import io.iworkflow.core.command.CommandResults;
import io.iworkflow.core.command.TimerCommand;
import io.iworkflow.core.communication.Communication;
import io.iworkflow.core.persistence.Persistence;

import java.time.Duration;

import static io.iworkflow.integ.basic.MixOfWithWaitUntilAndSkipWaitUntilWorkflow.SHARED_STATE_OPTIONS;

public class MixOfWithWaitUntilAndSkipWaitUntilState2 implements WorkflowState<Integer> {

    @Override
    public Class<Integer> getInputType() {
        return Integer.class;
    }

    @Override
    public CommandRequest waitUntil(
            final Context context,
            final Integer input,
            Persistence persistence,
            final Communication communication) {
        return CommandRequest.forAllCommandCompleted(TimerCommand.createByDuration(Duration.ofSeconds(1)));
    }

    @Override
    public StateDecision execute(
            final Context context,
            final Integer input,
            final CommandResults commandResults,
            Persistence persistence,
            final Communication communication) {
        final int output = input + 1;
        commandResults.getAllTimerCommandResults().get(0).getTimerStatus();
        return StateDecision.gracefulCompleteWorkflow(output);
    }

    @Override
    public WorkflowStateOptions getStateOptions() {
        return SHARED_STATE_OPTIONS;
    }
}