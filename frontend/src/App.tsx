import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AppLayout } from './components/layout/AppLayout'
import { LandingPage } from './pages/LandingPage'
import { DashboardPage } from './pages/DashboardPage'
import { ChatPage } from './pages/ChatPage'
import { InboxPage } from './pages/InboxPage'
import { EchoesPage } from './pages/EchoesPage'
import { CodexPage } from './pages/CodexPage'
import { NoteDetailPage } from './pages/NoteDetailPage'
import { IndexPage } from './pages/IndexPage'
import { GlossaryPage } from './pages/GlossaryPage'
import { SettingsPage } from './pages/SettingsPage'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route element={<AppLayout />}>
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="/chat" element={<ChatPage />} />
          <Route path="/inbox" element={<InboxPage />} />
          <Route path="/echoes" element={<EchoesPage />} />
          <Route path="/codex" element={<CodexPage />} />
          <Route path="/codex/:id" element={<NoteDetailPage />} />
          <Route path="/index" element={<IndexPage />} />
          <Route path="/glossary" element={<GlossaryPage />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
