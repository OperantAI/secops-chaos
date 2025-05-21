from pydantic import BaseModel

class AIExperimentVerifierResult(BaseModel):
    check: str
    entityType: str
    detected: bool
    score: float