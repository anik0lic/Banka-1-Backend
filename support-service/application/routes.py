from fastapi import APIRouter
from pydantic import BaseModel
from .chatbot import customer_service_response 
router = APIRouter()

class ChatMessage(BaseModel):
    message: str

@router.post("/")
def chat_endpoint(msg: ChatMessage):
    return {"response": customer_service_response(msg.message)}  