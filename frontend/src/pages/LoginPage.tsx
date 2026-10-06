import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

export default function LoginPage() {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState("");
    const { login } = useAuth();
    const navigate = useNavigate();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError("");
        try {
            const response = await login({ username, password });
            if (response.requiresTwoFactor) {
                navigate("/2fa/verify", { state: { challengeId: response.challengeId } });
            } else if (response.token) {
                navigate("/");
            }
        } catch (err: any) {
            const msg = err.response?.data?.error || err.response?.data?.message || "Login failed";
            setError(msg);
        }
    };

    return (
        <div className="max-w-md mx-auto mt-10 p-6 border rounded shadow">
            <h1 className="text-heading text-center">Login</h1>
            <form onSubmit={handleSubmit} className="space-y-4">
                <input
                    type="text"
                    placeholder="Username"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    required
                    className="w-full"
                />
                <input
                    type="password"
                    placeholder="Password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                    className="w-full"
                />
                {error && <div className="error-text">{error}</div>}
                <button type="submit" className="btn-primary w-full">Login</button>
            </form>
            <p className="mt-4 text-center">
                Don't have an account? <Link to="/register" className="text-primary">Register</Link>
            </p>
        </div>
    );
}