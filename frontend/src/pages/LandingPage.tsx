import { login } from '../api/auth'

export function LandingPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-zinc-950 text-zinc-100">
      <h1 className="mb-4 text-4xl font-bold">Cerebray</h1>
      <p className="mb-8 text-lg text-zinc-400">
        A personal Zettelkasten powered by AI conversations.
      </p>
      <button
        onClick={() => login()}
        className="rounded-md bg-white px-6 py-3 text-sm font-medium text-zinc-900 hover:bg-zinc-200 transition-colors"
      >
        Sign in
      </button>
    </div>
  )
}
