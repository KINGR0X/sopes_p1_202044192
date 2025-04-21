from pydantic import BaseModel
from typing import List, Optional
from datetime import datetime


class logContainer(BaseModel):
    pid: int
    name: str
    container_id: str
    memory_usage: float
    cpu_usage: float
    disk_usage_read_mb: float
    disk_usage_write_mb: float
    io_read_mb: float
    io_write_mb: float
    action: str
    timestamp: str
    creation_time: Optional[str] = None
    deletion_time: Optional[str] = None


class logMemory(BaseModel):
    total_ram_mb: float
    free_ram_mb: float
    usage_ram_mb: float
    timestamp: str
