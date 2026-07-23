# Copyright (c) 2022-2026 Super Durable, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from iwf.command_results import CommandResults
from iwf.communication import Communication
from iwf.persistence import Persistence
from iwf.persistence_schema import PersistenceField, PersistenceSchema
from iwf.state_decision import StateDecision
from iwf.state_schema import StateSchema
from iwf.workflow import ObjectWorkflow
from iwf.workflow_context import WorkflowContext
from iwf.workflow_state import WorkflowState

TEST_DA_KEY = "test-da"


class State1(WorkflowState[None]):
    def execute(
        self,
        ctx: WorkflowContext,
        input: None,
        command_results: CommandResults,
        persistence: Persistence,
        communication: Communication,
    ) -> StateDecision:
        assert input is None
        test_da = persistence.get_data_attribute(TEST_DA_KEY)
        assert test_da is None

        return StateDecision.graceful_complete_workflow(output="success")


class EmptyDataWorkflow(ObjectWorkflow):
    def get_workflow_states(self) -> StateSchema:
        return StateSchema.with_starting_state(State1())

    def get_persistence_schema(self) -> PersistenceSchema:
        return PersistenceSchema.create(
            PersistenceField.data_attribute_def(TEST_DA_KEY, None),
        )
