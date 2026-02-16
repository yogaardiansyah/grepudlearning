// app/dashboard/page.tsx
'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import axios from 'axios';

const gateway = axios.create({
  baseURL: 'http://localhost:8000',
});

// Setup Interceptor untuk pasang Token otomatis
gateway.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  console.log("Token yang dikirim:", token); // <--- TAMBAHKAN INI
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  } else {
    console.warn("Token tidak ditemukan di localStorage!");
  }
  return config;
});


interface Order {
  id: string;
  item: string;
  price: number;
  status: string;
}

export default function DashboardPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [newItem, setNewItem] = useState('Nasi Goreng');
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  // Load Orders
  const fetchOrders = async () => {
    try {
      const res = await gateway.get('/order/list');
      setOrders(res.data || []);
    } catch (err) {
      console.error('Gagal load order', err);
      // Jika 401 Unauthorized, lempar ke login
    //   router.push('/auth/login');
    }
  };

  useEffect(() => {
    fetchOrders();
  }, []);

  // Handle Buat Order
  const handleOrder = async () => {
    setLoading(true);
    try {
      await gateway.post('/order/create', {
        item: newItem,
        price: 25000,
      });
      fetchOrders(); // Refresh list
    } catch (err) {
      alert('Gagal membuat pesanan');
    } finally {
      setLoading(false);
    }
  };

  // Handle Bayar
  const handlePay = async (orderId: string, amount: number) => {
    if (!confirm('Bayar pesanan ini?')) return;
    try {
      await gateway.post('/payment/pay', {
        order_id: orderId,
        amount: amount,
      });
      alert('Pembayaran Berhasil!');
      fetchOrders(); // Refresh status jadi 'paid'
    } catch (err) {
      alert('Pembayaran Gagal');
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 p-10">
      <div className="max-w-4xl mx-auto bg-white p-8 rounded shadow">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-bold">Dashboard Pemesanan</h1>
          <button 
            onClick={() => { localStorage.clear(); router.push('/auth/login'); }}
            className="text-red-500 hover:underline">
            Logout
          </button>
        </div>

        {/* Form Order */}
        <div className="flex gap-4 mb-8 p-4 bg-blue-50 rounded">
          <select 
            value={newItem} 
            onChange={(e) => setNewItem(e.target.value)}
            className="p-2 border rounded flex-1"
          >
            <option value="Nasi Goreng">Nasi Goreng - Rp 25.000</option>
            <option value="Ayam Bakar">Ayam Bakar - Rp 30.000</option>
            <option value="Es Teh Manis">Es Teh Manis - Rp 5.000</option>
          </select>
          <button 
            onClick={handleOrder} 
            disabled={loading}
            className="bg-blue-600 text-white px-6 py-2 rounded hover:bg-blue-700"
          >
            {loading ? 'Memproses...' : 'Pesan Sekarang'}
          </button>
        </div>

        {/* List Order */}
        <h2 className="text-xl font-semibold mb-4">Riwayat Pesanan</h2>
        <div className="space-y-4">
          {orders.length === 0 && <p className="text-gray-500">Belum ada pesanan.</p>}
          
          {orders.map((order) => (
            <div key={order.id} className="border p-4 rounded flex justify-between items-center">
              <div>
                <p className="font-bold text-lg">{order.item}</p>
                <p className="text-gray-600">Harga: Rp {order.price.toLocaleString()}</p>
                <p className={`text-sm font-semibold mt-1 ${order.status === 'paid' ? 'text-green-600' : 'text-yellow-600'}`}>
                  Status: {order.status.toUpperCase()}
                </p>
              </div>
              
              {order.status === 'pending' && (
                <button
                  onClick={() => handlePay(order.id, order.price)}
                  className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700"
                >
                  Bayar
                </button>
              )}
              {order.status === 'paid' && (
                <span className="bg-gray-200 text-gray-600 px-4 py-2 rounded cursor-default">
                  Lunas
                </span>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}