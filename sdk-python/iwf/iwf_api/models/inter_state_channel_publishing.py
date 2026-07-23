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

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.encoded_object import EncodedObject


T = TypeVar("T", bound="InterStateChannelPublishing")


@_attrs_define
class InterStateChannelPublishing:
    """
    Attributes:
        channel_name (str):
        value (Union[Unset, EncodedObject]):
    """

    channel_name: str
    value: Union[Unset, "EncodedObject"] = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        channel_name = self.channel_name

        value: Union[Unset, dict[str, Any]] = UNSET
        if not isinstance(self.value, Unset):
            value = self.value.to_dict()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "channelName": channel_name,
            }
        )
        if value is not UNSET:
            field_dict["value"] = value

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.encoded_object import EncodedObject

        d = dict(src_dict)
        channel_name = d.pop("channelName")

        _value = d.pop("value", UNSET)
        value: Union[Unset, EncodedObject]
        if isinstance(_value, Unset):
            value = UNSET
        else:
            value = EncodedObject.from_dict(_value)

        inter_state_channel_publishing = cls(
            channel_name=channel_name,
            value=value,
        )

        inter_state_channel_publishing.additional_properties = d
        return inter_state_channel_publishing

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
