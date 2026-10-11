import React, { useState } from 'react'
import { NetFusionProvider } from './context/NetFusionContext'
import { Header } from './components/layout/Header'
import { Sidebar, type NavTab } from './components/layout/Sidebar'
import { ChaosModal } from './components/common/ChaosModal'
import { DashboardPage } from './pages/DashboardPage'
import { NetworksPage } from './pages/NetworksPage'
import { TrafficPage } from './pages/TrafficPage'
import { DownloadsPage } from './pages/DownloadsPage'
import { BenchmarkPage } from './pages/BenchmarkPage'
import { HistoryPage } from './pages/HistoryPage'
import { PoliciesPage } from './pages/PoliciesPage'
import { SettingsPage } from './pages/SettingsPage'

const AppContent: React.FC = () => {
  const [currentTab, setCurrentTab] = useState<NavTab>('dashboard')
  const [isChaosOpen, setIsChaosOpen] = useState(false)

  const renderPage = () => {
    switch (currentTab) {
      case 'dashboard':
        return <DashboardPage />
      case 'networks':
        return <NetworksPage />
      case 'traffic':
        return <TrafficPage />
      case 'downloads':
        return <DownloadsPage />
      case 'benchmark':
        return <BenchmarkPage />
      case 'history':
        return <HistoryPage />
      case 'policies':
        return <PoliciesPage />
      case 'settings':
        return <SettingsPage />
      default:
        return <DashboardPage />
    }
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      {/* Top Sticky Header */}
      <Header onOpenChaos={() => setIsChaosOpen(true)} />

      {/* Main Container with Sidebar + Content */}
      <div style={{ display: 'flex', flex: 1 }}>
        <Sidebar currentTab={currentTab} onSelectTab={setCurrentTab} />

        <main
          style={{
            flex: 1,
            padding: '28px 36px 60px',
            maxWidth: '1600px',
            margin: '0 auto',
            width: '100%',
          }}
        >
          {renderPage()}
        </main>
      </div>

      {/* Chaos Injection Modal */}
      <ChaosModal isOpen={isChaosOpen} onClose={() => setIsChaosOpen(false)} />
    </div>
  )
}

export function App() {
  return (
    <NetFusionProvider>
      <AppContent />
    </NetFusionProvider>
  )
}

export default App
