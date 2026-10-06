import { Outlet, Link } from "react-router-dom";
import { useAuth } from "../../contexts/AuthContext";

export default function Layout() {
    const { token, logout } = useAuth();
    const { user } = useAuth();

    return (
        <div className="min-h-screen flex flex-col">
            <header className="bg-primary text-white p-4">
                <div className="container mx-auto flex justify-between items-center">
                    <Link to="/" className="text-heading font-bold">GeoGuide</Link>
                    <nav className="space-x-4">
                        <Link to="/countries" className="hover:underline">Countries</Link>
                        {token && (
                            <>
                                <Link to="/trips" className="hover:underline">My Trips</Link>
                                <Link to="/reviews" className="hover:underline">My Reviews</Link>
                                <Link to="/profile" className="hover:underline">Profile</Link>

                                {(user?.role === 'MODERATOR' || user?.role === 'ANALYST') && (
                                    <Link to="/moderate/reviews" className="hover:underline">Moderate Reviews</Link>
                                )}
                                {user?.role === 'MODERATOR' && (
                                    <Link to="/moderate/users" className="hover:underline">Manage Users</Link>
                                )}
                                {user?.role === 'ANALYST' && (
                                    <Link to="/analytics" className="hover:underline">Analytics</Link>
                                )}
                                <Link to="/favourites" className="hover:underline">Favourites</Link>

                                <Link to="/profile" className="inline-flex items-center">
                                    <img src={user?.avatarUrl || '/default-avatar.png'} alt="Avatar" className="w-12 h-12 rounded-full object-cover" />
                                </Link>
                                <button onClick={logout} className="bg-red-600 px-3 py-1 rounded">Logout</button>
                            </>
                        )}
                        {!token && <Link to="/login" className="hover:underline">Login</Link>}
                    </nav>
                </div>
            </header>
            <main className="flex-grow container mx-auto p-4">
                <Outlet />
            </main>
            <footer className="bg-gray-100 text-center p-4 text-secondary">
                GeoGuide &copy; 2025
            </footer>
        </div>
    );
}