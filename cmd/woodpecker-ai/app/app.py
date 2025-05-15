from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from openai import OpenAI
from typing import List
from pydantic import BaseModel
from .verifiers import AIExperimentVerifierResult
from .sensitive_data_verifier import VerifySensitiveData
import os

def create_app() -> FastAPI:

    app = FastAPI(
        title="Woodpecker AI API",
    )

    register_routes(app)

    return app

client = OpenAI(
    api_key=os.getenv("OPENAI_KEY")
)

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
    @app.get("/")
    async def root():
        return {"message": "Hello World"}



    class AIExperiment(BaseModel):
        model: str
        ai_api: str
        system_prompt: str
        prompt: str
        verify_prompt_checks: List[str]
        verify_response_checks: List[str]

    class AIExperimentResponse(BaseModel):
        model: str
        ai_api: str
        prompt: str
        api_response: str
        verified_prompt_checks: List[AIExperimentVerifierResult]
        verified_response_checks: List[AIExperimentVerifierResult]


    @app.post("/ai-experiments")
    async def chat(experiment: AIExperiment):
        match experiment.model:
            case "gpt-4o":
                completion = client.chat.completions.create(
                    model="gpt-4o",
                    messages = [
                        {"role": "system", "content": experiment.system_prompt},
                        {"role": "user", "content": experiment.prompt}
                    ]
                )
                verified_prompt_checks = list()
                verified_response_checks = list()
                for check in experiment.verify_prompt_checks:
                    match check:
                        case "PII":
                            results = VerifySensitiveData(check, experiment.system_prompt + experiment.prompt)
                            for i in results:
                                verified_prompt_checks.append(i)
                for check in experiment.verify_response_checks:
                    match check:
                        case "PII":
                            results = VerifySensitiveData(check, completion.choices[0].message.content)
                            for i in results:
                                verified_response_checks.append(i)

                return AIExperimentResponse(model=experiment.model, ai_api=experiment.ai_api, prompt=experiment.prompt,
                                            api_response=completion.choices[0].message.content, verified_prompt_checks=verified_prompt_checks,
                                            verified_response_checks=verified_response_checks)
