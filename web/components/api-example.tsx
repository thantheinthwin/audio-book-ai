"use client";

import { useState, useEffect } from "react";
import { authAPI, profileAPI } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export function ApiExample() {
  const [user, setUser] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Example: Get current user profile
  const fetchUserProfile = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await profileAPI.getProfile();
      setUser(response.user);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to fetch profile");
    } finally {
      setLoading(false);
    }
  };

  // Example: Validate token
  const validateToken = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await authAPI.validateToken();
      console.log("Token validation response:", response);
      alert("Token is valid!");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Token validation failed");
    } finally {
      setLoading(false);
    }
  };

  // Example: Health check
  const checkHealth = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await authAPI.health();
      console.log("Health check response:", response);
      alert("API is healthy!");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Health check failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>API Integration Example</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex gap-2">
            <Button onClick={fetchUserProfile} disabled={loading}>
              {loading ? "Loading..." : "Get Profile"}
            </Button>

            <Button
              onClick={validateToken}
              disabled={loading}
              variant="outline"
            >
              Validate Token
            </Button>

            <Button onClick={checkHealth} disabled={loading} variant="outline">
              Health Check
            </Button>
          </div>

          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-red-600 text-sm">{error}</p>
            </div>
          )}

          {user && (
            <div className="p-4 bg-gray-50 border rounded-md">
              <h3 className="font-semibold mb-2">User Profile:</h3>
              <pre className="text-sm overflow-auto">
                {JSON.stringify(user, null, 2)}
              </pre>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
