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
import io.iworkflow.gen.models.WorkflowConfig;
import io.iworkflow.integ.basic.SkipWaitUntilWorkflow;
import io.iworkflow.spring.TestSingletonWorkerService;
import io.iworkflow.spring.controller.WorkflowRegistry;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.util.concurrent.ExecutionException;

public class SkipWaitUntilTest {

    @BeforeEach
    public void setup() throws ExecutionException, InterruptedException {
        TestSingletonWorkerService.startWorkerIfNotUp();
    }

    @Test
    public void testSkipWaitUntil() throws InterruptedException {
        final Client client = new Client(WorkflowRegistry.registry, ClientOptions.localDefault);
        final String wfId = "testSkipWaitUntil-" + System.currentTimeMillis() / 1000;
        final WorkflowOptions startOptions = ImmutableWorkflowOptions.builder()
                .workflowIdReusePolicy(IDReusePolicy.DISALLOW_REUSE)
                .workflowConfigOverride(new WorkflowConfig().continueAsNewThreshold(1))
                .build();
        final int input = 0;
        client.startWorkflow(SkipWaitUntilWorkflow.class, wfId, 10, input, startOptions);
        // wait for workflow to finish
        final Integer output = client.getSimpleWorkflowResultWithWait(Integer.class, wfId);
        Assertions.assertEquals(input + 2, output);
    }
}
