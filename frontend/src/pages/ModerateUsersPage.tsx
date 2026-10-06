import { useEffect, useState } from 'react';
import { getUsers, moderateUser } from '../api/auth';
import type { UserResponse } from '../types/api';

export default function ModerateUsersPage() {
    const [users, setUsers] = useState<UserResponse[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [message, setMessage] = useState('');

    const fetchUsers = async () => {
        setLoading(true);
        setError('');
        try {
            const data = await getUsers({ limit: 100 });
            setUsers(data);
        } catch (err: any) {
            console.error(err);
            setError(err.response?.data?.error || 'Failed to load users');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchUsers();
    }, []);

    const handleBlock = async (userId: string, block: boolean, reason?: string) => {
        try {
            await moderateUser(userId, { block, reason: reason || '' });
            setMessage(block ? 'User blocked' : 'User unblocked');
            fetchUsers();
            setTimeout(() => setMessage(''), 3000);
        } catch (err: any) {
            alert(err.response?.data?.error || 'Action failed');
        }
    };

    if (loading) return <div>Loading users...</div>;
    if (error) return <div className="error-text">{error}</div>;

    return (
        <div>
            <h1 className="text-heading">User Management</h1>
            {message && <div className="p-2 bg-green-100 text-green-800 rounded mb-4">{message}</div>}
            <div className="overflow-x-auto">
                <table className="min-w-full border">
                    <thead>
                        <tr>
                            <th>Username</th>
                            <th>Email</th>
                            <th>Registered</th>
                            <th>Status</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {users.map((user) => (
                            <tr key={user.id}>
                                <td>{user.username}</td>
                                <td>{user.email}</td>
                                <td>{user.registeredAt ? new Date(user.registeredAt).toLocaleDateString() : ''}</td>
                                <td>{user.isBlocked ? 'Blocked' : 'Active'}</td>
                                <td>
                                    {!user.isBlocked ? (
                                        <button
                                            onClick={() => {
                                                const reason = prompt('Enter block reason:');
                                                if (reason) handleBlock(user.id!, true, reason);
                                            }}
                                            className="btn-secondary"
                                        >
                                            Block
                                        </button>
                                    ) : (
                                        <button
                                            onClick={() => handleBlock(user.id!, false)}
                                            className="btn-primary"
                                        >
                                            Unblock
                                        </button>
                                    )}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
}