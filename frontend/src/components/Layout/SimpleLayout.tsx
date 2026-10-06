import { Outlet, Link } from "react-router-dom";

export default function SimpleLayout() {
    return (
        <div className="min-h-screen flex flex-col">
            <header className="bg-primary text-white p-4">
                <div className="container mx-auto flex justify-between items-center">
                    <Link to="/" className="text-heading font-bold">GeoGuide</Link>
                    <nav className="space-x-4">
                        <Link to="/countries" className="hover:underline">Countries</Link>
                        <Link to="/login" className="hover:underline">Login</Link>
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