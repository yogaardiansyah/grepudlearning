import Link from 'next/link';

export default function HomePage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-24">
      <div className="space-y-4 text-center">
        <h1 className="text-4xl font-bold">Selamat Datang</h1>
        <p className="text-gray-600">Aplikasi Autentikasi dengan Go + Next.js</p>
        <div className="flex justify-center gap-4 pt-4">
          <Link href="/auth/login" className="px-6 py-2 font-semibold text-white bg-blue-600 rounded-lg hover:bg-blue-700">
            Login
          </Link>
          <Link href="/auth/register" className="px-6 py-2 font-semibold text-gray-800 bg-gray-200 rounded-lg hover:bg-gray-300">
            Register
          </Link>
        </div>
      </div>
    </main>
  );
}