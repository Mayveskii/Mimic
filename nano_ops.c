#include &lt;stdio.h&gt;
#include &lt;stdlib.h&gt;
#include &lt;string.h&gt;

// Читает файл и возвращает содержимое как строку
char* native_read_file(const char* path) {
    FILE* file = fopen(path, "r");
    if (!file) return NULL;

    fseek(file, 0, SEEK_END);
    long size = ftell(file);
    rewind(file);

    char* data = (char*)malloc(size + 1);
    if (!data) {
        fclose(file);
        return NULL;
    }

    size_t read = fread(data, 1, size, file);
    fclose(file);
    data[read] = '\0';
    return data;
}

// Записывает строку в файл
int native_write_file(const char* path, const char* data) {
    FILE* file = fopen(path, "w");
    if (!file) return -1;

    fprintf(file, "%s", data);
    fclose(file);
    return 0;
}

// Выполняет diff между двумя файлами через системную команду
char* native_diff_files(const char* path1, const char* path2) {
    char command[512];
    snprintf(command, sizeof(command), "diff -u %s %s", path1, path2);

    FILE* pipe = popen(command, "r");
    if (!pipe) return NULL;

    char buffer[4096];
    char* result = (char*)malloc(65536); // 64K buffer
    if (!result) {
        pclose(pipe);
        return NULL;
    }
    result[0] = '\0';

    while (fgets(buffer, sizeof(buffer), pipe)) {
        strcat(result, buffer);
    }
    pclose(pipe);

    return result;
}
