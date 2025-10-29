function formatDateForInput(dateStr: string) {
  const date = new Date(dateStr); // JS парсит ISO/UTC
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0"); // месяцы с 0
  const day = String(date.getDate()).padStart(2, "0");
  return `${day}-${month}-${year}`;
}


export default formatDateForInput