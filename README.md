# short-url

Web service for short url

**Stack**
- React (Vite) frontend
- Go backend (Chi)
**Prereqs**
- Node.js 20+
- pnpm 10+
- Go 1.22+
**Install**
```bash
pnpm install
```

**Run**
Frontend dev server:
```bash
pnpm --filter frontend dev
```

Backend dev server:
```bash
pnpm --filter backend-chi dev
```
**Test**
```bash
pnpm -r run test
```

**Build**
```bash
pnpm -r run build
```

**Go Module**

The Go module path is:
```
github.com/danielxfeng/short-url/apps/backend-chi
```
