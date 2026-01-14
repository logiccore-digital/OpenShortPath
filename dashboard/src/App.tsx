import { BrowserRouter, Routes, Route } from "react-router-dom"
import { Toaster } from "sonner"
import { Dashboard } from "./pages/Dashboard"
import { Login } from "./pages/Login"
import { Signup } from "./pages/Signup"
import { ShortURLDetail } from "./pages/ShortURLDetail"
import { Account } from "./pages/Account"
import { Namespaces } from "./pages/Namespaces"
import { ProtectedRoute } from "./components/ProtectedRoute"
import { ThemeProvider } from "./components/theme/ThemeProvider"

function App() {
  return (
    <ThemeProvider>
      <BrowserRouter basename="/dashboard">
        <Toaster position="bottom-right" richColors />
        <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/signup" element={<Signup />} />
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
          path="/namespaces"
          element={
            <ProtectedRoute>
              <Namespaces />
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

