import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { createReview } from "../api/reviews";

export default function CreateReviewPage() {
    const { placeId } = useParams();
    const navigate = useNavigate();
    const [rating, setRating] = useState(5);
    const [comment, setComment] = useState("");
    const [visitDate, setVisitDate] = useState(new Date().toISOString().slice(0, 10));
    const [error, setError] = useState("");

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            await createReview(placeId!, { rating, comment, visitDate });
            navigate(`/places/${placeId}`);
        } catch (err: any) {
            setError(err.response?.data?.error || "Failed to submit review");
        }
    };

    return (
        <div className="max-w-xl mx-auto">
            <h1 className="text-heading">Write a Review</h1>
            <form onSubmit={handleSubmit} className="space-y-4 mt-4">
                <div>
                    <label className="block font-medium">Rating (1-5)</label>
                    <select value={rating} onChange={e => setRating(parseInt(e.target.value))} className="w-full">
                        {[1, 2, 3, 4, 5].map(r => <option key={r}>{r}</option>)}
                    </select>
                </div>
                <div>
                    <label className="block font-medium">Comment</label>
                    <textarea value={comment} onChange={e => setComment(e.target.value)} rows={5} required className="w-full" />
                </div>
                <div>
                    <label className="block font-medium">Visit Date</label>
                    <input type="date" value={visitDate} onChange={e => setVisitDate(e.target.value)} required className="w-full" />
                </div>
                {error && <div className="text-accent">{error}</div>}
                <button type="submit" className="btn-primary">Submit Review</button>
            </form>
        </div>
    );
}