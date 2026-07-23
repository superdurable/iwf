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

from enum import Enum


class ErrorSubStatus(str, Enum):
    LONG_POLL_TIME_OUT_SUB_STATUS = "LONG_POLL_TIME_OUT_SUB_STATUS"
    UNCATEGORIZED_SUB_STATUS = "UNCATEGORIZED_SUB_STATUS"
    WORKER_API_ERROR = "WORKER_API_ERROR"
    WORKFLOW_ALREADY_STARTED_SUB_STATUS = "WORKFLOW_ALREADY_STARTED_SUB_STATUS"
    WORKFLOW_NOT_EXISTS_SUB_STATUS = "WORKFLOW_NOT_EXISTS_SUB_STATUS"

    def __str__(self) -> str:
        return str(self.value)
