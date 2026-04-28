import ctypes
import os
import tempfile

# Компилируем C-код в разделяемую библиотеку
def compile_c_lib():
    if not os.path.exists("./libnano_ops.so"):
        os.system("gcc -fPIC -shared -o libnano_ops.so nano_ops.c")

compile_c_lib()

# Загружаем библиотеку
lib = ctypes.CDLL("./libnano_ops.so")

# Определяем сигнатуры функций
lib.native_read_file.argtypes = [ctypes.c_char_p]
lib.native_read_file.restype = ctypes.c_char_p

lib.native_write_file.argtypes = [ctypes.c_char_p, ctypes.c_char_p]
lib.native_write_file.restype = ctypes.c_int

lib.native_diff_files.argtypes = [ctypes.c_char_p, ctypes.c_char_p]
lib.native_diff_files.restype = ctypes.c_char_p

# Python-обёртки
def read_file(path: str) -> str:
    result = lib.native_read_file(path.encode())
    return result.decode() if result else None

def write_file(path: str, data: str) -> bool:
    return lib.native_write_file(path.encode(), data.encode()) == 0

def diff_files(path1: str, path2: str) -> str:
    result = lib.native_diff_files(path1.encode(), path2.encode())
    return result.decode() if result else ""

# Простой тест
if __name__ == "__main__":
    # Создадим тестовые файлы
    with open("file1.txt", "w") as f:
        f.write("Hello\nWorld\n")
    with open("file2.txt", "w") as f:
        f.write("Hello\nUniverse\n")

    # Протестируем diff
    diff = diff_files("file1.txt", "file2.txt")
    print("=== DIFF ===")
    print(diff)
