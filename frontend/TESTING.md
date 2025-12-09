# Frontend Testing Strategy

This document outlines the testing strategy for the Gitwapp frontend (React + TypeScript).

## 1. Unit & Integration Testing
**Tools:** `Vitest` + `React Testing Library` + `MSW (Mock Service Worker)`

- [x] **Infrastructure Setup**
    - [x] Install dependencies (`vitest`, `@testing-library/react`, `jsdom`, `msw`, `@testing-library/user-event`).
    - [x] Configure `vitest.config.ts`.
    - [x] Setup `src/setupTests.ts` (MSW server lifecycle, cleanup).

- [x] **Tests**
    - [x] **Login Flow (`Login.test.tsx`)**
        - [x] Renders form correctly.
        - [x] Handles input entry.
        - [x] Mocks successful login -> redirects to `/`.
        - [x] Mocks failed login -> displays error message.
    - [x] **Dashboard (`Dashboard.test.tsx`)**
        - [x] Renders loading state.
        - [x] Renders empty state.
        - [x] Renders list of repositories.
    - [x] **Repo Detail (`RepoDetail.test.tsx`)**
        - [x] Renders status (Branch, Clean/Dirty, Ahead/Behind).
        - [x] Renders list of changed files.
        - [x] "Stage" button triggers API call.
        - [x] "Commit" flow (input + button).

## 2. End-to-End (E2E) Testing
**Tools:** `Playwright`

- [ ] **Infrastructure Setup**
    - [ ] Install Playwright.
    - [ ] Configure Playwright to run against the Go backend (`./bin/server serve`).

- [ ] **Scenarios**
    - [ ] **Smoke Test:** Load page, Login, see Dashboard.

## 3. Static Analysis
- [ ] **Linting:** ESLint (already configured via Vite template).
- [ ] **Type Checking:** `tsc` (already configured).

---

## Progress Log

- **[Date]**: Document created.
