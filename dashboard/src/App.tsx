import { BrowserRouter, Routes, Route } from "react-router-dom"
import { Toaster } from "sonner"
import { Dashboard } from "./pages/Dashboard"
import { Login } from "./pages/Login"
import { ShortURLDetail } from "./pages/ShortURLDetail"
import { Account } from "./pages/Account"
import { ProtectedRoute } from "./components/ProtectedRoute"
import { ThemeProvider } from "./components/theme/ThemeProvider"

function App() {
  return (
    <ThemeProvider>
      <BrowserRouter basename="/dashboard">
        <Toaster position="bottom-right" richColors />
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
          path="/account"
          element={
            <ProtectedRoute>
              <Account />
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
    </ThemeProvider>
  )
}

export default App

