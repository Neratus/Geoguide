import { useEffect, useState } from 'react';
import { getPendingReviews, moderateReview } from '../api/reviews';
import type { ReviewBriefResponse } from '../types/api';

export default function ModerateReviewsPage() {
    const [reviews, setReviews] = useState<ReviewBriefResponse[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [message, setMessage] = useState('');

    const fetchReviews = async () => {
        setLoading(true);
        setError('');
        try {
            const data = await getPendingReviews({ limit: 50 });
            setReviews(data);
        } catch (err: any) {
            console.error(err);
            setError(err.response?.data?.error || 'Failed to load pending reviews');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchReviews();
    }, []);

    const handleModerate = async (reviewId: string, approved: boolean, comment?: string) => {
        try {

            await moderateReview(reviewId, { approved, comment });
            setMessage(approved ? 'Review approved' : 'Review rejected');
            fetchReviews();
            setTimeout(() => setMessage(''), 3000);
        } catch (err: any) {
            alert(err.response?.data?.error || 'Moderation failed');
        }
    };

    if (loading) return <div>Loading pending reviews...</div>;
    if (error) return <div className="error-text">{error}</div>;

    return (
        <div>
            <h1 className="text-heading">Moderate Reviews</h1>
            {message && <div className="p-2 bg-green-100 text-green-800 rounded mb-4">{message}</div>}
            <div className="space-y-4">
                {reviews.length === 0 && <p>No pending reviews.</p>}
                {reviews.map((rev) => (
                    <div key={rev.id} className="border p-4 rounded">
                        <div className="flex justify-between items-start">
                            <div>
                                <p className="font-semibold">{rev.username}</p>
                                <p>Rating: {rev.rating}/5</p>
                                <p className="text-secondary">{rev.comment}</p>
                                {rev.visitDate && <p>Visit: {rev.visitDate}</p>}
                                <p className="text-xs text-gray-500">Created: {rev.createdAt ? new Date(rev.createdAt).toLocaleString() : ''}</p>
                            </div>
                            <div className="space-x-2">
                                <button
                                    onClick={() => handleModerate(rev.id!, true)}
                                    className="btn-primary"
                                >
                                    Approve
                                </button>
                                <button
                                    onClick={() => handleModerate(rev.id!, false, 'Violates guidelines')}
                                    className="btn-secondary"
                                >
                                    Reject
                                </button>
                            </div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}