# NVIDIA setup for careTech

1. Create or sign in to a NVIDIA account with access to the hosted model endpoint.
2. Generate an API key in the NVIDIA developer or model access area.
3. Copy `.env.example` to `.env`.
4. Fill in the real values for:
   - `NVIDIA_API_KEY`
   - `NVIDIA_API_BASE_URL`
   - `NVIDIA_MODEL`
5. Run the app:

   PowerShell:
   ```powershell
   Copy-Item .env.example .env
   $env:NVIDIA_API_KEY = (Get-Content .env | Select-String 'NVIDIA_API_KEY=').Line.Split('=')[1]
   $env:NVIDIA_API_BASE_URL = (Get-Content .env | Select-String 'NVIDIA_API_BASE_URL=').Line.Split('=')[1]
   $env:NVIDIA_MODEL = (Get-Content .env | Select-String 'NVIDIA_MODEL=').Line.Split('=')[1]
   go run ./cmd/server
   ```

6. Test the chat endpoint:

   ```powershell
   $body = '{"session_id":"demo-nvidia","message":"Есть ABB-S201-C16?"}'
   Invoke-RestMethod -Uri 'http://localhost:8080/api/chat' -Method Post -ContentType 'application/json' -Body $body
   ```

Important:
- Never commit `.env` to Git.
- Keep NVIDIA credentials in the local environment or a secure secrets manager.
- The app still uses the local catalog as the source of truth for product facts.
