import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

export default function TwoFactorPage() {
    const [code, setCode] = useState("");
    const [error, setError] = useState("");
    const { verifyTwoFactor } = useAuth();
    const location = useLocation();
    const navigate = useNavigate();
    const challengeId = location.state?.challengeId;

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!challengeId) {
            setError("Invalid session");
            return;
        }
        try {
            await verifyTwoFactor(challengeId, code);
            navigate("/");
        } catch (err: any) {
            setError(err.response?.data?.error || "Invalid code");
        }
    };

    return (
        <div className="max-w-md mx-auto mt-10 p-6 border rounded shadow">
            <h1 className="text-heading text-center">Two‑Factor Authentication</h1>
            <form onSubmit={handleSubmit} className="space-y-4">
                <input
                    type="text"
                    placeholder="8‑digit code"
                    value={code}
                    onChange={(e) => setCode(e.target.value)}
                    maxLength={8}
                    required
                    className="w-full text-center text-2xl"
                />
                {error && <div className="error-text">{error}</div>}
                <button type="submit" className="btn-primary w-full">Verify</button>
            </form>
        </div>
    );
}