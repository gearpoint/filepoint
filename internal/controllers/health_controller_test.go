package controllers_test

// func TestHealthRoute(t *testing.T) {
// 	s := server.NewServer(&config.Config{
// 		Server: config.ServerConfig{Addr: "80"},
// 		AWS:    config.AWSConfig{},
// 		Kafka:  config.KafkaConfig{},
// 		Redis:  config.RedisConfig{},
// 	}, nil, nil)

// 	w := httptest.NewRecorder()
// 	req, err := http.NewRequest("GET", "v1/health", nil)
// 	assert.Nil(t, err)

// 	s.Engine.ServeHTTP(w, req)

// 	assert.Equal(t, 200, w.Code)
// 	assert.Equal(t, "OK", w.Body.String())
// }
