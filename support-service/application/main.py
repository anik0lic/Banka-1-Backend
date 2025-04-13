from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from .routes import router
import os
import logging


log_level = os.getenv("LOG_LEVEL", "info").upper()
logging.basicConfig(level=log_level)

app = FastAPI(
    title="Banking Chatbot API",
    description="API for interacting with the banking support chatbot",
    version="0.1.0"
)

@app.get("/health")
async def health_check():
    return {"status": "healthy"}

# CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(router, prefix="/chat")