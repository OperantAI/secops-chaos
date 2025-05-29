from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from openai import OpenAI
from pydantic import BaseModel
import os
import structlog

LOGGER = structlog.getLogger(__name__)

def create_app() -> FastAPI:

    app = FastAPI(
        title="Woodpecker AI App",
    )

    register_routes(app)

    return app

client = OpenAI(api_key=os.getenv("OPENAI_KEY"))

class ChatRequest(BaseModel):
    model: str
    system_prompt: str
    prompt: str

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

    @app.post("/chat")
    async def chat(chat_request: ChatRequest):
        completion = client.chat.completions.create(
            model=chat_request.model,
            messages=[
                {"role": "system", "content": chat_request.system_prompt},
                {
                    "role": "user",
                    "content": chat_request.prompt,
                },
            ],
        )
        LOGGER.info(f"response from OpenAI {completion.choices[0].message.content}")
        return {"message": completion.choices[0].message.content}

    @app.on_event("shutdown")
    async def shutdown_event():
        LOGGER.info("Shutting down app...")
