from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from openai import OpenAI
import os

def create_app() -> FastAPI:

    app = FastAPI(
        title="Woodpecker AI App",
    )

    register_routes(app)

    return app

client = OpenAI(api_key=os.getenv("OPENAI_KEY"))

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
    async def root():
        completion = client.chat.completions.create(
            model="gpt-4o",
            messages=[
                {"role": "system", "content": "You are a helpful banking assistant"},
                {
                    "role": "user",
                    "content": "Can you give me information regarding my account?",
                },
            ],
        )
        return {"message": completion}
