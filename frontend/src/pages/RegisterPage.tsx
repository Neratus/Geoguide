import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

const validateEmail = (email: string): string | null => {
    const re = /^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$/i;
    if (!email) return "Email is required";
    if (!re.test(email)) return "Invalid email format";
    return null;
};

const validatePassword = (password: string): string | null => {
    if (password.length < 8) return "Password must be at least 8 characters";
    if (!/[A-Z]/.test(password)) return "Password must contain at least one uppercase letter";
    if (!/[a-z]/.test(password)) return "Password must contain at least one lowercase letter";
    if (!/[0-9]/.test(password)) return "Password must contain at least one number";
    return null;
};


export default function RegisterPage() {
    const [username, setUsername] = useState("");
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState("");
    const [success, setSuccess] = useState("");
    const { register } = useAuth();
    const navigate = useNavigate();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError("");
        const emailError = validateEmail(email);
        if (emailError) { setError(emailError); return; }

        const passwordError = validatePassword(password);
        if (passwordError) { setError(passwordError); return; }
        try {
            const result = await register({ username, email, password });
            if (result?.id) {
                navigate(`/verify?userId=${result.id}&contact=${encodeURIComponent(email)}`);
            } else {
                setSuccess("Registration successful! Please check your email for verification code.");
                setTimeout(() => navigate("/login"), 3000);
            }
        } catch (err: any) {
            setError(err.response?.data?.error || "Registration failed");
        }
    };

    return (
        <div className="max-w-md mx-auto mt-10 p-6 border rounded shadow">
            <h1 className="text-heading text-center">Register</h1>
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
                    type="email"
                    placeholder="Email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
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
                {success && <div className="text-green-600 text-sm">{success}</div>}
                <button type="submit" className="btn-primary w-full">Register</button>
            </form>
            <p className="mt-4 text-center">
                Already have an account? <Link to="/login" className="text-primary">Login</Link>
            </p>
        </div>
    );
}