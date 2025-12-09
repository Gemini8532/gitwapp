# Frontend Testing Strategy

This document outlines the testing strategy for the Gitwapp frontend (React + TypeScript).

## 1. Unit & Integration Testing
**Tools:** `Vitest` + `React Testing Library` + `MSW (Mock Service Worker)`

- [ ] **Infrastructure Setup**
    - [ ] Install dependencies (`vitest`, `@testing-library/react`, `jsdom`, `msw`, `@testing-library/user-event`).
    - [ ] Configure `vitest.config.ts`.
    - [ ] Setup `src/setupTests.ts` (MSW server lifecycle, cleanup).

- [ ] **Tests**
    - [ ] **Login Flow (`Login.test.tsx`)**
        - [ ] Renders form correctly.
        - [ ] Handles input entry.
        - [ ] Mocks successful login -> redirects to `/`.
        - [ ] Mocks failed login -> displays error message.
    - [ ] **Dashboard (`Dashboard.test.tsx`)**
        - [ ] Renders loading state.
        - [ ] Renders empty state.
        - [ ] Renders list of repositories.
    - [ ] **Repo Detail (`RepoDetail.test.tsx`)**
        - [ ] Renders status (Branch, Clean/Dirty, Ahead/Behind).
        - [ ] Renders list of changed files.
        - [ ] "Stage" button triggers API call.
        - [ ] "Commit" flow (input + button).

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
