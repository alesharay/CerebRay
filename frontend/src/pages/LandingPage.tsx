import { login } from '../api/auth'
import { MessageSquare, BookOpen, Network, Sparkles } from 'lucide-react'

const features = [
  {
    icon: MessageSquare,
    title: 'Learn through conversation',
    desc: 'Chat with AI about any topic. Ask questions, explore ideas, and build understanding naturally.',
  },
  {
    icon: Sparkles,
    title: 'AI-structured notes',
    desc: 'Conversations turn into Zettelkasten notes - summaries, analogies, core ideas, and connections created for you.',
  },
  {
    icon: BookOpen,
    title: 'Grow your knowledge base',
    desc: 'Notes flow from Inbox to Echoes to Codex as your understanding deepens. Promote what matters, let the rest rest.',
  },
  {
    icon: Network,
    title: 'Connect everything',
    desc: 'Link related ideas across topics. See how concepts relate through connections and graph views.',
  },
]

export function LandingPage() {
  return (
    <div className="flex min-h-screen flex-col bg-zinc-950 text-zinc-100">
      {/* Hero */}
      <div className="flex flex-1 flex-col items-center justify-center px-6 py-24">
        <h1 className="mb-3 text-5xl font-bold tracking-tight">Cereb-Ray</h1>
        <p className="mb-8 max-w-md text-center text-lg text-zinc-400">
          A personal Zettelkasten powered by AI conversations. Learn, capture, connect.
        </p>
        <button
          onClick={() => login()}
          className="rounded-lg bg-white px-8 py-3 text-sm font-semibold text-zinc-900 transition-colors hover:bg-zinc-200"
        >
          Sign in to get started
        </button>
      </div>

      {/* Features */}
      <div className="border-t border-zinc-900 px-6 py-16">
        <div className="mx-auto grid max-w-4xl grid-cols-1 gap-8 sm:grid-cols-2">
          {features.map((f) => (
            <div key={f.title} className="flex gap-4">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-zinc-900">
                <f.icon className="h-5 w-5 text-zinc-400" />
              </div>
              <div>
                <h3 className="mb-1 font-medium">{f.title}</h3>
                <p className="text-sm text-zinc-500">{f.desc}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
