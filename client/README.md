# URL Shortener - Client App

React + TypeScript + Vite application for the URL shortener dashboard.

## Architecture

This client app is designed to run on a subdomain (`app.my-domain.com`) to separate concerns:

- **Marketing Site**: `my-domain.com` (Astro) - Reserved for marketing website
- **App Dashboard**: `app.my-domain.com` (This React app) - User dashboard and link management
- **Short Links**: `my-domain.com/[shortcode]` - Handled by backend redirect service

This architecture avoids routing conflicts and provides clean separation of concerns.

## Setup

### 1. Install Dependencies

```bash
pnpm install
```

### 2. Environment Variables

Create a `.env` file in the client directory:

```bash
cp .env.example .env
```

Then add your Clerk publishable key:

```env
VITE_CLERK_PUBLISHABLE_KEY=pk_test_...
VITE_CLERK_SIGN_IN_URL=/login
VITE_CLERK_SIGN_UP_URL=/login
VITE_CLERK_AFTER_SIGN_IN_URL=/
VITE_CLERK_AFTER_SIGN_UP_URL=/
```

Get your Clerk keys from [Clerk Dashboard](https://dashboard.clerk.com).

### 3. Configure Clerk Dashboard for OAuth-Only Authentication

Since this app only supports Google and GitHub OAuth (no email/password), you need to configure this in the Clerk Dashboard:

1. Go to [Clerk Dashboard](https://dashboard.clerk.com) → Your Application
2. Navigate to **User & Authentication** → **Social Connections**
3. Enable **Google** and **GitHub** OAuth providers
4. Navigate to **User & Authentication** → **Email, Phone, Username**
5. **Disable** Email/Password authentication (or leave it disabled)

This ensures users can only sign in/sign up using Google or GitHub. With OAuth, sign-in and sign-up are the same flow - new users will automatically be created on first OAuth login.

### 4. Run Development Server

```bash
pnpm dev
```

## Routes

- `/` - Main dashboard showing all user links (protected)
- `/[shortcode]` - Individual link detail view (protected)
- `/login` - Clerk authentication page

## Tech Stack

- **React 19** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **TanStack Router** - File-based routing
- **Clerk** - Authentication
- **Tailwind CSS** - Styling
- **TanStack Query** - Data fetching
- **Zustand** - State management

## Project Structure

```
src/
├── routes/           # File-based routes (TanStack Router)
│   ├── __root.tsx   # Root layout with ClerkProvider
│   ├── index.tsx    # Home route - Links dashboard
│   ├── $shortcode.tsx  # Link detail page
│   └── login.tsx    # Login page
├── components/      # Reusable components
│   ├── ui/          # shadcn/ui components
│   └── ProtectedRoute.tsx  # Auth wrapper component
├── lib/             # Utilities
└── router.tsx       # Router configuration
```

---

## Original Template Info

This template provides a minimal setup to get React working in Vite with HMR and some ESLint rules.

Currently, two official plugins are available:

- [@vitejs/plugin-react](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react) uses [Babel](https://babeljs.io/) (or [oxc](https://oxc.rs) when used in [rolldown-vite](https://vite.dev/guide/rolldown)) for Fast Refresh
- [@vitejs/plugin-react-swc](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react-swc) uses [SWC](https://swc.rs/) for Fast Refresh

## React Compiler

The React Compiler is enabled on this template. See [this documentation](https://react.dev/learn/react-compiler) for more information.

Note: This will impact Vite dev & build performances.

## Expanding the ESLint configuration

If you are developing a production application, we recommend updating the configuration to enable type-aware lint rules:

```js
export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // Other configs...

      // Remove tseslint.configs.recommended and replace with this
      tseslint.configs.recommendedTypeChecked,
      // Alternatively, use this for stricter rules
      tseslint.configs.strictTypeChecked,
      // Optionally, add this for stylistic rules
      tseslint.configs.stylisticTypeChecked,

      // Other configs...
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // other options...
    },
  },
])
```

You can also install [eslint-plugin-react-x](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-x) and [eslint-plugin-react-dom](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-dom) for React-specific lint rules:

```js
// eslint.config.js
import reactX from 'eslint-plugin-react-x'
import reactDom from 'eslint-plugin-react-dom'

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // Other configs...
      // Enable lint rules for React
      reactX.configs['recommended-typescript'],
      // Enable lint rules for React DOM
      reactDom.configs.recommended,
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // other options...
    },
  },
])
```
