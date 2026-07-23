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

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, Union

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..models.workflow_error_type import WorkflowErrorType
from ..models.workflow_status import WorkflowStatus
from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.state_completion_output import StateCompletionOutput


T = TypeVar("T", bound="WorkflowGetResponse")


@_attrs_define
class WorkflowGetResponse:
    """
    Attributes:
        workflow_run_id (str):
        workflow_status (WorkflowStatus):
        results (Union[Unset, list['StateCompletionOutput']]):
        error_type (Union[Unset, WorkflowErrorType]):
        error_message (Union[Unset, str]):
    """

    workflow_run_id: str
    workflow_status: WorkflowStatus
    results: Union[Unset, list["StateCompletionOutput"]] = UNSET
    error_type: Union[Unset, WorkflowErrorType] = UNSET
    error_message: Union[Unset, str] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        workflow_run_id = self.workflow_run_id

        workflow_status = self.workflow_status.value

        results: Union[Unset, list[dict[str, Any]]] = UNSET
        if not isinstance(self.results, Unset):
            results = []
            for results_item_data in self.results:
                results_item = results_item_data.to_dict()
                results.append(results_item)

        error_type: Union[Unset, str] = UNSET
        if not isinstance(self.error_type, Unset):
            error_type = self.error_type.value

        error_message = self.error_message

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "workflowRunId": workflow_run_id,
                "workflowStatus": workflow_status,
            }
        )
        if results is not UNSET:
            field_dict["results"] = results
        if error_type is not UNSET:
            field_dict["errorType"] = error_type
        if error_message is not UNSET:
            field_dict["errorMessage"] = error_message

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.state_completion_output import StateCompletionOutput

        d = dict(src_dict)
        workflow_run_id = d.pop("workflowRunId")

        workflow_status = WorkflowStatus(d.pop("workflowStatus"))

        results = []
        _results = d.pop("results", UNSET)
        for results_item_data in _results or []:
            results_item = StateCompletionOutput.from_dict(results_item_data)

            results.append(results_item)

        _error_type = d.pop("errorType", UNSET)
        error_type: Union[Unset, WorkflowErrorType]
        if isinstance(_error_type, Unset):
            error_type = UNSET
        else:
            error_type = WorkflowErrorType(_error_type)

        error_message = d.pop("errorMessage", UNSET)

        workflow_get_response = cls(
            workflow_run_id=workflow_run_id,
            workflow_status=workflow_status,
            results=results,
            error_type=error_type,
            error_message=error_message,
        )

        workflow_get_response.additional_properties = d
        return workflow_get_response

    @property
    def additional_keys(self) -> list[str]:
        return list(self.additional_properties.keys())

    def __getitem__(self, key: str) -> Any:
        return self.additional_properties[key]

    def __setitem__(self, key: str, value: Any) -> None:
        self.additional_properties[key] = value

    def __delitem__(self, key: str) -> None:
        del self.additional_properties[key]

    def __contains__(self, key: str) -> bool:
        return key in self.additional_properties
