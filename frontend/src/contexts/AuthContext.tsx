import { createContext, useContext, useState, useEffect, ReactNode } from "react";
import { login, register, verifyTwoFactor, getProfile, logout } from "../api/auth";
import type { LoginRequest, RegisterRequest, ProfileResponse } from "../types/api";

interface AuthContextType {
    user: ProfileResponse | null;
    token: string | null;
    isLoading: boolean;
    login: (data: LoginRequest) => Promise<{ token?: string; requiresTwoFactor?: boolean; challengeId?: string }>;
    register: (data: RegisterRequest) => Promise<void>;
    verifyTwoFactor: (challengeId: string, code: string) => Promise<void>;
    logout: () => Promise<void>;
    refreshProfile: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
    const [user, setUser] = useState<ProfileResponse | null>(null);
    const [token, setToken] = useState<string | null>(localStorage.getItem("token"));
    const [isLoading, setIsLoading] = useState(true);

    const refreshProfile = async () => {
        if (!token) return;
        try {
            const response = await getProfile();
            setUser(response.data);
        } catch (error) {
            console.error("Failed to load profile", error);
            localStorage.removeItem("token");
            setToken(null);
            setUser(null);
        }
    };

    useEffect(() => {
        if (token) {
            refreshProfile().finally(() => setIsLoading(false));
            // setIsLoading(false);
        } else {
            setIsLoading(false);
        }
    }, [token]);

    const handleLogin = async (data: LoginRequest): Promise<{ token?: string; requiresTwoFactor?: boolean; challengeId?: string }> => {
        const response = await login(data);
        const { token, requiresTwoFactor, challengeId } = response.data;
        if (token) {
            localStorage.setItem("token", token);
            setToken(token);
            await refreshProfile();
            return { token };
        }
        return { requiresTwoFactor: true, challengeId: challengeId! };
    };

    const handleRegister = async (data: RegisterRequest) => {
        const response = await register(data);
        return response.data;
    };

    const handleVerifyTwoFactor = async (challengeId: string, code: string) => {
        const response = await verifyTwoFactor({ challengeId, code });
        const newToken = response.data.token;
        localStorage.setItem("token", newToken);
        setToken(newToken);
        await refreshProfile();
    };

    const handleLogout = async () => {
        try {
            await logout();
        } catch (error) {
            console.error("Logout error", error);
        }
        localStorage.removeItem("token");
        setToken(null);
        setUser(null);
    };

    return (
        <AuthContext.Provider
            value={{
                user,
                token,
                isLoading,
                login: handleLogin,
                register: handleRegister,
                verifyTwoFactor: handleVerifyTwoFactor,
                logout: handleLogout,
                refreshProfile,
            }}
        >
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) throw new Error("useAuth must be used within AuthProvider");
    return context;
};