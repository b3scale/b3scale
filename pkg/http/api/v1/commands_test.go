package v1

import "testing"

func TestBackendMeetingsEnd(t *testing.T) {
	/*
		u, _ := url.Parse("http:///?backend_host=" + backend.Backend.Host)
		req := &http.Request{
			URL: u,
		}
		ctx, rec := MakeTestContext(req)
		defer ctx.Release()
		ctx = AuthorizeTestContext(ctx, "admin42", []string{ScopeAdmin})

		if err := BackendMeetingsEnd(ctx); err != nil {
			t.Fatal(err)
		}
		res := rec.Result()
		if res.StatusCode != http.StatusAccepted {
			t.Error("unexpected status code:", res.StatusCode)
		}
		resBody, _ := ioutil.ReadAll(res.Body)
		t.Log("list:", string(resBody))
	*/

}
