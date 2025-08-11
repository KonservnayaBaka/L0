import React, { useState } from "react";

export default function OrderInfo() {
  const [orderUID, setOrderUID] = useState("");
  const [orderData, setOrderData] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  const fetchOrder = async () => {
    setError(null);
    setOrderData(null);

    if (!orderUID.trim()) {
      setError("Please enter orderUID");
      return;
    }

    setLoading(true);
    try {
      const res = await fetch(`http://localhost:8081/order/${orderUID}`);
      if (!res.ok) {
        const errData = await res.json();
        throw new Error(errData.error || "Failed to fetch order");
      }
      const data = await res.json();
      setOrderData(data);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
      <div style={{ maxWidth: 700, margin: "20px auto", fontFamily: "Arial, sans-serif" }}>
        <h1>Order Info Lookup</h1>
        <input
            type="text"
            placeholder="Enter orderUID"
            value={orderUID}
            onChange={(e) => setOrderUID(e.target.value)}
            style={{ padding: 8, width: "60%", fontSize: 16 }}
        />
        <button onClick={fetchOrder} style={{ marginLeft: 10, padding: "8px 16px", fontSize: 16 }}>
          Fetch Order
        </button>

        {loading && <p>Loading...</p>}

        {error && <p style={{ color: "red" }}>{error}</p>}

        {orderData && (
            <div style={{ marginTop: 20, border: "1px solid #ccc", padding: 15, borderRadius: 5 }}>
              <h2>Order UID: {orderData.orderUID}</h2>
              <p><strong>Track Number:</strong> {orderData.trackNumber}</p>
              <p><strong>Entry:</strong> {orderData.entry}</p>
              <p><strong>Locale:</strong> {orderData.locale}</p>
              <p><strong>Customer ID:</strong> {orderData.customerID}</p>
              <h3>Delivery</h3>
              <p>Name: {orderData.delivery.name}</p>
              <p>Phone: {orderData.delivery.phone}</p>
              <p>Zip: {orderData.delivery.zip}</p>
              <p>City: {orderData.delivery.city}</p>
              <p>Address: {orderData.delivery.address}</p>
              <p>Region: {orderData.delivery.region}</p>
              <p>Email: {orderData.delivery.email}</p>

              <h3>Payment</h3>
              <p>Transaction: {orderData.payment.transaction}</p>
              <p>Amount: {orderData.payment.amount}</p>
              <p>Currency: {orderData.payment.currency}</p>
              <p>Provider: {orderData.payment.provider}</p>
              <p>Payment Date: {new Date(orderData.payment.paymentDt * 1000).toLocaleString()}</p>

              <h3>Items</h3>
              <ul>
                {orderData.items.map((item) => (
                    <li key={item.chrtID} style={{ marginBottom: 10 }}>
                      <strong>{item.name}</strong> (Brand: {item.brand})<br />
                      Price: {item.price}, Sale: {item.sale}, Size: {item.size}, Total Price: {item.totalPrice}
                    </li>
                ))}
              </ul>
            </div>
        )}
      </div>
  );
}