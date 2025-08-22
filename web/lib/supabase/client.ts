import { createBrowserClient } from "@supabase/ssr";

export function createClient() {
  return createBrowserClient(
    "https://eyjgtkmnpnunwrqmnoae.supabase.co",
    "sb_publishable_hWoNGiZRTHbu05qwzlMCRg_ETFwifo8"
  );
}
