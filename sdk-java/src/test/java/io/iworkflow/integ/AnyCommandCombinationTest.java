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
import io.iworkflow.core.WorkflowUncompletedException;
import io.iworkflow.gen.models.WorkflowErrorType;
import io.iworkflow.gen.models.WorkflowStatus;
import io.iworkflow.integ.anycommandcombination.AnyCommandCombinationFailWorkflow;
import io.iworkflow.spring.TestSingletonWorkerService;
import io.iworkflow.spring.controller.WorkflowRegistry;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.util.concurrent.ExecutionException;

class AnyCommandCombinationTest {

    @BeforeEach
    public void setup() throws ExecutionException, InterruptedException {
        TestSingletonWorkerService.startWorkerIfNotUp();
    }

    @Test
    void testStateApiFailWorkflow() {
        final Client client = new Client(WorkflowRegistry.registry, ClientOptions.localDefault);
        final long startTs = System.currentTimeMillis();
        final String wfId = "testStateApiFailWorkflow" + startTs / 1000;
        final Integer input = 5;

        final String runId = client.startWorkflow(
                AnyCommandCombinationFailWorkflow.class, wfId, 10, input);

        try {
            client.waitForWorkflowCompletion(Integer.class, wfId);
        } catch (WorkflowUncompletedException e) {
            Assertions.assertEquals(runId, e.getRunId());
            Assertions.assertEquals(WorkflowStatus.FAILED, e.getClosedStatus());
            Assertions.assertEquals(WorkflowErrorType.STATE_API_FAIL_ERROR_TYPE, e.getErrorSubType());
            Assertions.assertTrue(e.getErrorMessage().contains("CommandNotFoundException: Found unknown commandId in the combination list"));
            Assertions.assertEquals(0, e.getStateResultsSize());
            return;
        }
        Assertions.fail("no exception caught");
    }
}
