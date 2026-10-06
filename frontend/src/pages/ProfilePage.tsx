import { useEffect, useState } from "react";
import { useAuth } from "../contexts/AuthContext";
import { updateProfile } from "../api/auth";
import apiClient from "../api/client";
import type { ProfileResponse } from "../types/api";

export default function ProfilePage() {
    const { user, refreshProfile } = useAuth();
    const [formData, setFormData] = useState<Partial<ProfileResponse>>({});
    const [editing, setEditing] = useState(false);
    const [message, setMessage] = useState("");
    const [uploading, setUploading] = useState(false);

    useEffect(() => {
        if (user) {
            setFormData({
                username: user.username,
                phone: user.phone,
                birthDate: user.birthDate,
                countryOfResidence: user.countryOfResidence,
                favoriteCategories: user.favoriteCategories,
            });
        }
    }, [user]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            await updateProfile(formData);
            await refreshProfile();
            setMessage("Profile updated successfully");
            setEditing(false);
            setTimeout(() => setMessage(""), 3000);
        } catch (err: any) {
            setMessage(err.response?.data?.error || "Update failed");
        }
    };

    const handleAvatarUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        const formData = new FormData();
        formData.append("avatar", file);
        setUploading(true);
        setMessage("");
        try {
            await apiClient.post("api/v1/me/avatar", formData, {
                headers: { "Content-Type": "multipart/form-data" },
            });
            await refreshProfile();
            setMessage("Avatar updated successfully");
        } catch (err: any) {
            setMessage(err.response?.data?.error || "Avatar upload failed");
        } finally {
            setUploading(false);
            e.target.value = "";
        }
    };

    if (!user) return <div className="text-center p-8">Loading...</div>;

    return (
        <div className="max-w-2xl mx-auto">
            <h1 className="text-heading">My Profile</h1>
            {message && (
                <div className={message.includes("success") ? "text-green-600" : "text-red-600"}>
                    {message}
                </div>
            )}

            {!editing ? (
                <div className="space-y-4">
                    <div className="flex justify-center mb-4 relative">
                        <img
                            src={user.avatarUrl || "/default-avatar.png"}
                            alt="Avatar"
                            className="w-24 h-24 rounded-full object-cover"
                        />
                        <label className="absolute bottom-0 right-1/3 bg-blue-600 text-white text-xs px-2 py-1 rounded cursor-pointer">
                            {uploading ? "..." : "Change"}
                            <input
                                type="file"
                                accept="image/jpeg,image/png,image/webp"
                                onChange={handleAvatarUpload}
                                disabled={uploading}
                                className="hidden"
                            />
                        </label>
                    </div>
                    <div><span className="font-medium">Username:</span> {user.username}</div>
                    <div><span className="font-medium">Email:</span> {user.email}</div>
                    <div><span className="font-medium">Phone:</span> {user.phone || "—"}</div>
                    <div><span className="font-medium">Country of residence:</span> {user.countryOfResidence || "—"}</div>
                    <div><span className="font-medium">Birth date:</span> {user.birthDate || "—"}</div>
                    <div><span className="font-medium">Registered at:</span> {new Date(user.registeredAt).toLocaleDateString()}</div>
                    <div><span className="font-medium">Role:</span> {user.role}</div>
                    <button onClick={() => setEditing(true)} className="btn-primary">Edit Profile</button>
                </div>
            ) : (
                <form onSubmit={handleSubmit} className="space-y-4">
                    <input
                        type="text"
                        placeholder="Username"
                        value={formData.username || ""}
                        onChange={e => setFormData({ ...formData, username: e.target.value })}
                        className="w-full"
                    />
                    <input
                        type="text"
                        placeholder="Phone"
                        value={formData.phone || ""}
                        onChange={e => setFormData({ ...formData, phone: e.target.value })}
                        className="w-full"
                    />
                    <input
                        type="date"
                        placeholder="Birth date"
                        value={formData.birthDate?.split("T")[0] || ""}
                        onChange={e => setFormData({ ...formData, birthDate: e.target.value })}
                        className="w-full"
                    />
                    <input
                        type="text"
                        placeholder="Country of residence"
                        value={formData.countryOfResidence || ""}
                        onChange={e => setFormData({ ...formData, countryOfResidence: e.target.value })}
                        className="w-full"
                    />
                    <div className="flex gap-2">
                        <button type="submit" className="btn-primary">Save</button>
                        <button type="button" onClick={() => setEditing(false)} className="btn-secondary">Cancel</button>
                    </div>
                </form>
            )}
        </div>
    );
}