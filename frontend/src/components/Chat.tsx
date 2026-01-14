import { useState, useEffect, useRef } from 'react'
import './Chat.css'

interface Message {
  id: number
  createdAt: string
  connID: string
  user: string
  text: string
}

export default function Chat() {
  const [username, setUsername] = useState('anonimus')
  const [messages, setMessages] = useState<Message[]>([])
  const [inputText, setInputText] = useState('')
  const [isConnected, setIsConnected] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const reconnectTimeoutRef = useRef<number | null>(null)

  useEffect(() => {
    setMessages([])
    connectWebSocket()
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      if (wsRef.current) {
        wsRef.current.close()
        wsRef.current = null
      }
    }
  }, [username])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const connectWebSocket = () => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      return
    }

    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws?user=${encodeURIComponent(username)}`
    const ws = new WebSocket(wsUrl)
    
    ws.onopen = () => {
      setIsConnected(true)
      wsRef.current = ws
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
        reconnectTimeoutRef.current = null
      }
    }

    ws.onmessage = (event) => {
      try {
        const message: Message = JSON.parse(event.data)
        setMessages((prev) => {
          const exists = prev.some((m) => m.id === message.id)
          if (exists) {
            return prev
          }
          return [...prev, message]
        })
      } catch (error) {
        console.error('Failed to parse message:', error)
      }
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      setIsConnected(false)
    }

    ws.onclose = () => {
      setIsConnected(false)
      wsRef.current = null
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      reconnectTimeoutRef.current = setTimeout(() => {
        connectWebSocket()
      }, 3000)
    }
  }

  const sendMessage = () => {
    if (!inputText.trim() || !wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      return
    }

    try {
      wsRef.current.send(JSON.stringify({ text: inputText }))
      setInputText('')
    } catch (error) {
      console.error('Failed to send message:', error)
    }
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      sendMessage()
    }
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    })
  }

  return (
    <div className="chat-container">
      <div className="chat-header">
        <div className="username-section">
          <label htmlFor="username">Имя пользователя:</label>
          <input
            id="username"
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            className="username-input"
          />
        </div>
        <div className={`connection-status ${isConnected ? 'connected' : 'disconnected'}`}>
          {isConnected ? 'Подключено' : 'Отключено'}
        </div>
      </div>

      <div className="messages-container">
        {messages.map((msg) => (
          <div key={msg.id} className="message">
            <div className="message-header">
              <span className="message-user">{msg.user}</span>
              <span className="message-date">{formatDate(msg.createdAt)}</span>
            </div>
            <div className="message-text">{msg.text}</div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      <div className="input-section">
        <input
          type="text"
          value={inputText}
          onChange={(e) => setInputText(e.target.value)}
          onKeyPress={handleKeyPress}
          placeholder="Введите сообщение..."
          className="message-input"
          disabled={!isConnected}
        />
        <button
          onClick={sendMessage}
          disabled={!isConnected || !inputText.trim()}
          className="send-button"
        >
          Отправить
        </button>
      </div>
    </div>
  )
}
