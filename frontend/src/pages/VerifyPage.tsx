import { useRef } from "react";
import { useState, useEffect } from "react";
import { useSearchParams, useNavigate } from "react-router-dom";
import { verifyContact } from "../api/auth";


export default function VerifyPage() {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const [code, setCode] = useState("");
    const [error, setError] = useState("");
    const [success, setSuccess] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);
    const isSubmittingRef = useRef(false);
    const userId = searchParams.get("userId") || "";
    const contact = searchParams.get("contact") || "";

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (isSubmittingRef.current) return;
        isSubmittingRef.current = true;
        setIsSubmitting(true);
        setError("");
        setSuccess("");
        try {
            await verifyContact({ userId, contact, code });
            setSuccess("Email verified! You can now log in.");
            setTimeout(() => navigate("/login"), 2000);
        } catch (err: any) {
            const msg = err.response?.data?.error || err.response?.data?.message || "Verification failed";
            setError(msg);
            isSubmittingRef.current = false;
            setIsSubmitting(false);
        }
    };

    return (
        <div className="max-w-md mx-auto mt-10 p-6 border rounded">
            <h1 className="text-heading text-center">Verify Email</h1>
            <form onSubmit={handleSubmit} className="space-y-4 mt-4">
                <input
                    type="text"
                    placeholder="Verification code"
                    value={code}
                    onChange={e => setCode(e.target.value)}
                    required
                    disabled={isSubmitting}
                    className="w-full text-center text-2xl"
                />
                {error && <div className="text-red-600">{error}</div>}
                {success && <div className="text-green-600">{success}</div>}
                <button
                    type="submit"
                    disabled={isSubmitting}
                    className="btn-primary w-full disabled:opacity-50"
                >
                    {isSubmitting ? "Verifying..." : "Verify"}
                </button>
            </form>
        </div>
    );
}