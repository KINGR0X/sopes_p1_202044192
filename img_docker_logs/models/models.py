from pydantic import BaseModel
from typing import List


class logContainer(BaseModel):
    pid: int
    name: str
    container_id: str

    vsz: int
    rss: int
    memory_usage: float
    cpu_usage: float
    action: str

    timestamp: str


class logMemory(BaseModel):
    total_ram: int
    free_ram: int
    usage_ram: int
    timestamp: str
