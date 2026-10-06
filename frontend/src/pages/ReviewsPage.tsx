import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { getUserReviews, deleteReview } from "../api/reviews";
import type { Review } from "../types/api";

export default function ReviewsPage() {
    const [reviews, setReviews] = useState<Review[]>([]);
    const [loading, setLoading] = useState(true);

    const fetchReviews = async () => {
        try {
            const res = await getUserReviews();
            setReviews(res.data);
        } catch (err) {
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => { fetchReviews(); }, []);

    const handleDelete = async (id: string) => {
        if (!confirm("Delete this review?")) return;
        try {
            await deleteReview(id);
            setReviews(reviews.filter(r => r.id !== id));
        } catch (err) {
            alert("Failed to delete");
        }
    };

    if (loading) return <div>Loading...</div>;

    return (
        <div>
            <h1 className="text-heading">My Reviews</h1>
            {reviews.length === 0 ? (
                <p>You haven't written any reviews yet.</p>
            ) : (
                <div className="space-y-4 mt-4">
                    {reviews.map(r => (
                        <div key={r.id} className="border p-3 rounded">
                            <div className="flex justify-between">
                                <div>
                                    <Link to={`/places/${r.placeId}`} className="font-medium text-primary">{r.placeName}</Link>
                                    <span className="ml-2">⭐ {r.rating}</span>
                                    <span className="ml-2 text-secondary text-sm">{r.visitDate}</span>
                                    {!r.isApproved && <span className="ml-2 text-accent text-sm">(pending moderation)</span>}
                                </div>
                                <button onClick={() => handleDelete(r.id)} className="text-accent">Delete</button>
                            </div>
                            <p className="mt-1">{r.comment}</p>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}