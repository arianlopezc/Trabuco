package com.acme.shop.jobs;

import com.acme.shop.order.OrderRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

@Component
public class DailyOrderReportJob {

    private static final Logger log = LoggerFactory.getLogger(DailyOrderReportJob.class);

    @Autowired
    private OrderRepository orderRepository;

    @Scheduled(cron = "0 0 6 * * *")
    public void emitDailyReport() {
        long total = orderRepository.count();
        log.info("Daily order report: total orders so far = {}", total);
    }
}
