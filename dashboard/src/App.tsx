import { BrowserRouter, Routes, Route } from "react-router-dom"
import { Dashboard } from "./pages/Dashboard"
import { Login } from "./pages/Login"
import { ShortURLDetail } from "./pages/ShortURLDetail"
import { ProtectedRoute } from "./components/ProtectedRoute"

function App() {
  return (
    <BrowserRouter basename="/dashboard">
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <Dashboard />
            </ProtectedRoute>
          }
        />
        <Route
          path="/short-urls/:id"
          element={
            <ProtectedRoute>
              <ShortURLDetail />
            </ProtectedRoute>
          }
        />
        <Route
          path="*"
          element={
            <ProtectedRoute>
              <Dashboard />
            </ProtectedRoute>
          }
        />
      </Routes>
    </BrowserRouter>
  )
}

export default App

