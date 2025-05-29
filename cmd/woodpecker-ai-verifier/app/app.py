from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from typing import List
from pydantic import BaseModel
from .verifiers import AIExperimentVerifierResult
from .sensitive_data_verifier import VerifySensitiveData
import structlog

LOGGER = structlog.getLogger(__name__)


class AIExperiment(BaseModel):
    model: str
    ai_api: str
    system_prompt: str
    prompt: str
    response: str
    verify_prompt_checks: List[str]
    verify_response_checks: List[str]


class AIExperimentResponse(BaseModel):
    model: str
    ai_api: str
    prompt: str
    api_response: str
    verified_prompt_checks: List[AIExperimentVerifierResult]
    verified_response_checks: List[AIExperimentVerifierResult]


def create_app() -> FastAPI:

    app = FastAPI(
        title="Woodpecker AI Verifier API",
    )

    register_routes(app)

    return app


def register_routes(
    app: FastAPI,
):
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["Authorization", "Content-Type"],
    )

    @app.get("/healthz")
    async def healthz():
        return {"status": "ok"}

    @app.post("/v1/ai-experiments")
    async def ai_experiment(experiment: AIExperiment):
        verified_prompt_checks = list()
        verified_response_checks = list()
        for check in experiment.verify_prompt_checks:
            match check:
                case "PII":
                    results = VerifySensitiveData(
                        check, experiment.system_prompt + experiment.prompt
                    )
                    for i in results:
                        verified_prompt_checks.append(i)
        for check in experiment.verify_response_checks:
            match check:
                case "PII":
                    results = VerifySensitiveData(
                        check, experiment.response,
                    )
                    for i in results:
                        verified_response_checks.append(i)

        LOGGER.info(f"Verified prompt checks: {verified_prompt_checks}")
        LOGGER.info(f"Verified response checks: {verified_response_checks}")

        return AIExperimentResponse(
            model=experiment.model,
            ai_api=experiment.ai_api,
            prompt=experiment.prompt,
            api_response=experiment.response,
            verified_prompt_checks=verified_prompt_checks,
            verified_response_checks=verified_response_checks,
        )

    @app.on_event("shutdown")
    async def shutdown_event():
        LOGGER.info("Shutting down app...")
