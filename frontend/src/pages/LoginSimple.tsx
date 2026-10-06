export default function LoginSimple() {
    return (
        <div className="max-w-md mx-auto mt-10 p-6 border rounded shadow">
            <h1 className="text-heading text-center">Login (Simple)</h1>
            <form className="space-y-4">
                <input type="text" placeholder="Username" className="w-full" />
                <input type="password" placeholder="Password" className="w-full" />
                <button type="submit" className="btn-primary w-full">Login</button>
            </form>
        </div>
    );
}