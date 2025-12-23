import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"

export function Placeholder() {
  return (
    <div className="min-h-screen bg-background p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-4xl font-bold mb-8">Dashboard</h1>
        <Card>
          <CardHeader>
            <CardTitle>Welcome to OpenShortPath Dashboard</CardTitle>
            <CardDescription>
              This is a placeholder page. The dashboard is ready for development.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <p className="text-muted-foreground">
                The dashboard is now set up with React, TypeScript, React Router, Tailwind CSS, and shadcn UI.
              </p>
              <div className="flex gap-2">
                <Button>Get Started</Button>
                <Button variant="outline">Learn More</Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

