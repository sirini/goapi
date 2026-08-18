#include <stdio.h>
#include <vips/vips.h>

// 메모리에서 JPEG를 만들고 다시 읽어 WebP로 변환해 주요 이미지 경로를 확인합니다.
int main(int argc, char **argv) {
  VipsImage *source = NULL;
  VipsImage *loaded = NULL;
  void *jpeg = NULL;
  void *webp = NULL;
  size_t jpeg_length = 0;
  size_t webp_length = 0;
  int status = 1;

  if (VIPS_INIT(argv[0])) {
    vips_error_exit(NULL);
  }
  if (vips_black(&source, 64, 64, "bands", 3, NULL) ||
      vips_jpegsave_buffer(source, &jpeg, &jpeg_length, NULL) ||
      vips_jpegload_buffer(jpeg, jpeg_length, &loaded, NULL) ||
      vips_webpsave_buffer(loaded, &webp, &webp_length, NULL)) {
    fprintf(stderr, "%s\n", vips_error_buffer());
    goto cleanup;
  }
  if (jpeg_length == 0 || webp_length == 0) {
    fprintf(stderr, "이미지 변환 결과가 비어 있습니다.\n");
    goto cleanup;
  }
  status = 0;

cleanup:
  g_free(webp);
  g_clear_object(&loaded);
  g_free(jpeg);
  g_clear_object(&source);
  vips_shutdown();
  return status;
}
