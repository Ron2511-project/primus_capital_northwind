import { Routes, Route } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import CustomerDetail from './pages/CustomerDetail'

export default function App() {
  return (
    <>
      <header className="app-header">
        <div>
          <h1>Northwind Collect</h1>
          <p>Priorización de cobranza — ¿dónde poner foco hoy?</p>
        </div>
      </header>
      <main>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/customers/:id" element={<CustomerDetail />} />
        </Routes>
      </main>
    </>
  )
}
