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

package io.iworkflow.integ;

import io.iworkflow.core.Client;
import io.iworkflow.core.ClientOptions;
import io.iworkflow.core.ImmutableWorkflowOptions;
import io.iworkflow.core.WorkflowOptions;
import io.iworkflow.gen.models.IDReusePolicy;
import io.iworkflow.integ.stateoptionsoverride.StateOptionsOverrideWorkflow;
import io.iworkflow.spring.TestSingletonWorkerService;
import io.iworkflow.spring.controller.WorkflowRegistry;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.util.concurrent.ExecutionException;

public class StateOptionsOverrideTest {
    @BeforeEach
    public void setup() throws ExecutionException, InterruptedException {
        TestSingletonWorkerService.startWorkerIfNotUp();
    }
    @Test
    public void testStateOptionsOverrideWorkflow() throws InterruptedException {
        final Client client = new Client(WorkflowRegistry.registry, ClientOptions.localDefault);
        final String wfId = "state-options-override-test-id" + System.currentTimeMillis() / 1000;
        final WorkflowOptions startOptions = ImmutableWorkflowOptions.builder()
                .workflowIdReusePolicy(IDReusePolicy.DISALLOW_REUSE)
                .build();
        final String input = "input";
        client.startWorkflow(StateOptionsOverrideWorkflow.class, wfId, 10, input, startOptions);
        // wait for workflow to finish
        final String output = client.getSimpleWorkflowResultWithWait(String.class, wfId);
        Assertions.assertEquals("input_state1_start_state1_decide_state2_start_state2_decide", output);
    }
}
