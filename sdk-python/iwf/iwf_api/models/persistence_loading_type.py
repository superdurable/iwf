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


class PersistenceLoadingType(str, Enum):
    LOAD_ALL_WITHOUT_LOCKING = "LOAD_ALL_WITHOUT_LOCKING"
    LOAD_ALL_WITH_PARTIAL_LOCK = "LOAD_ALL_WITH_PARTIAL_LOCK"
    LOAD_NONE = "LOAD_NONE"
    LOAD_PARTIAL_WITHOUT_LOCKING = "LOAD_PARTIAL_WITHOUT_LOCKING"
    LOAD_PARTIAL_WITH_EXCLUSIVE_LOCK = "LOAD_PARTIAL_WITH_EXCLUSIVE_LOCK"

    def __str__(self) -> str:
        return str(self.value)
