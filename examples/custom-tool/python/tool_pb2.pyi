from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ToolsRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ToolsResponse(_message.Message):
    __slots__ = ("definitions",)
    DEFINITIONS_FIELD_NUMBER: _ClassVar[int]
    definitions: _containers.RepeatedCompositeFieldContainer[Definition]
    def __init__(self, definitions: _Optional[_Iterable[_Union[Definition, _Mapping]]] = ...) -> None: ...

class Definition(_message.Message):
    __slots__ = ("name", "description", "parameters")
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    PARAMETERS_FIELD_NUMBER: _ClassVar[int]
    name: str
    description: str
    parameters: str
    def __init__(self, name: _Optional[str] = ..., description: _Optional[str] = ..., parameters: _Optional[str] = ...) -> None: ...

class ExecuteRequest(_message.Message):
    __slots__ = ("name", "parameters")
    NAME_FIELD_NUMBER: _ClassVar[int]
    PARAMETERS_FIELD_NUMBER: _ClassVar[int]
    name: str
    parameters: str
    def __init__(self, name: _Optional[str] = ..., parameters: _Optional[str] = ...) -> None: ...

class ResultResponse(_message.Message):
    __slots__ = ("data",)
    DATA_FIELD_NUMBER: _ClassVar[int]
    data: str
    def __init__(self, data: _Optional[str] = ...) -> None: ...
