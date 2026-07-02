const Exception500 = () => {
  return (
    <div className="flex items-center justify-center h-screen">
      <div className="text-center">
        <h1 className="text-6xl font-bold text-gray-300">500</h1>
        <p className="text-xl text-gray-500 mt-4">Internal Server Error</p>
      </div>
    </div>
  );
};

export default Exception500;
