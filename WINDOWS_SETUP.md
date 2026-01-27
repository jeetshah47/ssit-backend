# Windows Setup Guide - Fixing CGO Compilation Error

## Problem
You're encountering this error:
```
cc1.exe: sorry, unimplemented: 64-bit mode not compiled in
```

This happens when you have a 32-bit C compiler but a 64-bit Go installation.

## Solution 1: Install 64-bit C Compiler (Recommended)

### Option A: Install TDM-GCC (Easiest)
1. Download TDM-GCC 64-bit from: https://jmeubank.github.io/tdm-gcc/
2. Install it (default location: `C:\TDM-GCC-64`)
3. Add to PATH: `C:\TDM-GCC-64\bin`
4. Restart your terminal/PowerShell
5. Verify: `gcc --version` should show 64-bit

### Option B: Install MSYS2 with MinGW-w64
1. Download MSYS2 from: https://www.msys2.org/
2. Install and open MSYS2 terminal
3. Run: `pacman -S mingw-w64-x86_64-gcc`
4. Add to PATH: `C:\msys64\mingw64\bin`
5. Restart your terminal

### Option C: Install MinGW-w64 directly
1. Download from: https://www.mingw-w64.org/downloads/
2. Extract and add `bin` folder to PATH
3. Restart terminal

## Solution 2: Use Docker (Alternative)

If you prefer not to install a C compiler, you can run the backend in Docker:

1. Create a `Dockerfile` in `ssit-backend/`:
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

2. Update `docker-compose.yml` to include backend service
3. Run: `docker-compose up`

## Solution 3: Quick Test (Temporary Workaround)

**Note**: This won't work for PostgreSQL/MongoDB, but you can test other parts:

```powershell
$env:CGO_ENABLED=0
go run .\cmd\server\main.go
```

This disables CGO, but your database drivers will fail.

## Verify Installation

After installing a 64-bit compiler, verify:

```powershell
gcc --version
# Should show something like: gcc (TDM-GCC-64) 10.x.x

go env CGO_ENABLED
# Should show: CGO_ENABLED=1
```

## Recommended: TDM-GCC

For Windows, **TDM-GCC 64-bit** is the easiest and most reliable option:
- Single installer
- No complex configuration
- Works out of the box with Go

Download: https://github.com/jmeubank/tdm-gcc/releases

After installation, restart your terminal and try running the backend again.

